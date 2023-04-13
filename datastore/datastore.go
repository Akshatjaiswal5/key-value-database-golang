package datastore

import (
	"errors"
	"sync"
	"time"
)

type queue struct {
	mu     sync.Mutex
	values []string
}

type timedValue struct {
	mu         sync.Mutex
	Value      string
	expiryTime time.Time
}

type Datastore struct {
	mu     sync.RWMutex
	values map[string]*timedValue
	queue  map[string]*queue
}

func NewDatastore() *Datastore {
	return &Datastore{
		values: make(map[string]*timedValue),
		queue:  make(map[string]*queue),
	}
}

func (d *Datastore) Set(key string, value string, expirySeconds int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	newTimedValue := &timedValue{
		Value:      value,
		expiryTime: time.Date(2099, 12, 31, 23, 59, 59, 999999999, time.UTC),
	}

	if expirySeconds != 0 {
		newTimedValue.expiryTime = time.Now().Add(time.Duration(expirySeconds) * time.Second)
	}

	d.values[key] = newTimedValue
	return nil
}

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

func (d *Datastore) QPush(key string, values ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	q, ok := d.queue[key]
	if !ok {
		q = &queue{}
		d.queue[key] = q
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.values = append(q.values, values...)
	return nil
}

func (d *Datastore) QPop(key string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	q, ok := d.queue[key]
	if !ok {
		return "", errors.New("queue not found")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.values) == 0 {
		return "", errors.New("queue is empty")
	}

	value := q.values[len(q.values)-1]
	q.values = q.values[:len(q.values)-1]
	return value, nil
}

func (d *Datastore) BQPop(key string, timeout time.Duration) (string, error) {
	d.mu.Lock()
	q, ok := d.queue[key]
	if !ok {
		q = &queue{}
		d.queue[key] = q
	}
	d.mu.Unlock()

	q.mu.Lock()
	if len(q.values) > 0 {
		value := q.values[len(q.values)-1]
		q.values = q.values[:len(q.values)-1]
		q.mu.Unlock()
		return value, nil
	}

	timeoutChan := make(chan bool, 1)
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

	select {
	case <-timeoutChan:
		q.mu.Unlock()
		return "", errors.New("queue is empty")
	default:
	}

	valueChan := make(chan string, 1)
	q.mu.Unlock()

	select {
	case value := <-valueChan:
		return value, nil
	case <-timeoutChan:
		return "", errors.New("queue is empty")
	}
}
