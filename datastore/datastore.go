package datastore

import (
	"errors"
	"sync"
	"time"
)

// queue represents a simple queue.
type queue struct {
	mu         sync.Mutex
	values     []string
	valuesChan chan string // added a channel to communicate the popped value to BQPop
}

// timedValue represents a value that is stored in the datastore
// along with its expiry time.
type timedValue struct {
	mu         sync.Mutex
	Value      string
	expiryTime time.Time
}

// Datastore represents a simple in-memory datastore.
type Datastore struct {
	mu     sync.RWMutex
	values map[string]*timedValue
	queues map[string]*queue
}

// NewDatastore creates a new instance of Datastore.
func NewDatastore() *Datastore {
	return &Datastore{
		values: make(map[string]*timedValue),
		queues: make(map[string]*queue),
	}
}

// Set sets the value for the given key in the datastore.
func (d *Datastore) Set(key string, value string, expirySeconds int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	newTimedValue := &timedValue{
		Value: value,
	}

	if expirySeconds != 0 {
		newTimedValue.expiryTime = time.Now().Add(time.Duration(expirySeconds) * time.Second)
	} else {
		newTimedValue.expiryTime = time.Date(2099, 12, 31, 23, 59, 59, 999999999, time.UTC)
	}

	d.values[key] = newTimedValue
	return nil
}

// Get retrieves the value for the given key from the datastore.
func (d *Datastore) Get(key string) (*timedValue, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, ok := d.values[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	if value.expiryTime.Before(time.Now()) {
		delete(d.values, key)
		return nil, errors.New("key not found")
	}

	return value, nil
}

// QPush adds the given values to the end of the queue with the given key in the datastore.
// If the queue does not exist, it is created.
// This function is thread-safe.
func (d *Datastore) QPush(key string, values ...string) error {
	// Acquire a write lock on the datastore to ensure thread safety.
	d.mu.Lock()
	defer d.mu.Unlock()
	// Retrieve the queue with the given key from the datastore.
	q, ok := d.queues[key]

	// If the queue does not exist, create a new queue and add it to the datastore.
	if !ok {
		q = &queue{valuesChan: make(chan string)} // added a valuesChan to the queue struct
		d.queues[key] = q
	}

	// Acquire a write lock on the queue to ensure thread safety.
	q.mu.Lock()
	defer q.mu.Unlock()

	// Append the given values to the end of the queue.
	q.values = append(q.values, values...)

	if len(q.values) == 1 { // if valuesChan is waiting for a value, send it
		select {
		case q.valuesChan <- q.values[0]:
			q.values = q.values[1:]
		default:
		}
	}
	return nil
}

// QPop removes and returns the last element of the queue with the given key.
// If the queue does not exist or is empty, an error is returned.
func (d *Datastore) QPop(key string) (string, error) {
	// Lock the mutex to ensure thread safety.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get the queue with the given key.
	q, ok := d.queues[key]
	if !ok {
		return "", errors.New("queue not found")
	}

	// Lock the queue's mutex to ensure thread safety.
	q.mu.Lock()
	defer q.mu.Unlock()

	// If the queue is empty, return an error.
	if len(q.values) == 0 {
		return "", errors.New("queue is empty")
	}

	// Remove and return the last value in the queue.
	value := q.values[len(q.values)-1]
	q.values = q.values[:len(q.values)-1]

	if len(q.values) > 0 && q.valuesChan != nil { // if valuesChan is waiting for a value, send it
		select {
		case q.valuesChan <- q.values[0]:
			q.values = q.values[1:]
		default:
		}
	}

	return value, nil
}

func (d *Datastore) BQPop(key string, timeout time.Duration) (string, error) {
	// Lock the mutex for the datastore to ensure thread-safety
	d.mu.Lock()

	// Check if a queue for the given key exists, otherwise create a new one
	q, ok := d.queues[key]
	if !ok {
		q = &queue{}
		d.queues[key] = q
	}

	// Release the lock on the datastore
	d.mu.Unlock()

	// Lock the mutex for the queue to ensure thread-safety
	q.mu.Lock()

	// If the queue is not empty, pop the last value and return it
	if len(q.values) > 0 {
		value := q.values[len(q.values)-1]
		q.values = q.values[:len(q.values)-1]
		q.mu.Unlock()
		return value, nil
	}

	// Create a channel to signal timeout
	timeoutChan := make(chan bool, 1)

	// If timeout is greater than zero, create a timer and start it in a separate goroutine
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		go func() {
			select {
			case <-timer.C:
				timeoutChan <- true
			}
		}()
	}

	// Release the lock on the queue before blocking on a channel
	q.mu.Unlock()

	select {
	// If timeout is triggered, return an error
	case <-timeoutChan:
		return "", errors.New("queue is empty")

	// Otherwise, wait for a value to be pushed onto the channel
	case value := <-q.valuesChan:
		return value, nil
	}
}
