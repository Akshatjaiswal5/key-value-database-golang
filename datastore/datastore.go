package datastore

import (
	"errors"
	"sync"
	"time"
)

// queue represents a simple queue.
type queue struct {
	mu     sync.Mutex
	Values chan string // added a channel to store values
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

// QPush pushes given values to the channel with the given key in the datastore.
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
		q = &queue{Values: make(chan string, 100)}
		d.queues[key] = q
	}

	// Acquire a write lock on the queue to ensure thread safety.
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, v := range values {
		q.Values <- v // Put each value in the channel
	}

	return nil
}

// QPop pops an element from the channel
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

	select {
	case value := <-q.Values:
		return value, nil
	default:
		return "", errors.New("queue is empty")
	}
}

func (d *Datastore) BQPop(key string, timeout time.Duration) (string, error) {

	// Lock the mutex to ensure thread safety.
	d.mu.Lock()

	// Get the queue with the given key.
	q, ok := d.queues[key]
	if !ok {
		q = &queue{Values: make(chan string, 100)} // Check if a queue for the given key exists, otherwise create a new one
		d.queues[key] = q
	}

	// Create a channel to signal timeout
	timeoutChan := make(chan bool, 1)

	//create a timer and start it in a separate goroutine
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			timeoutChan <- true
		}
	}()

	// Release the lock on the datastore before blocking on a channel
	d.mu.Unlock()

	select {
	// If timeout is triggered, return an error
	case <-timeoutChan:
		return "", errors.New("queue is empty")

	// Otherwise, wait for a value to be pushed onto the channel
	case value := <-q.Values:
		return value, nil
	}
}
