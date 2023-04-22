package datastore

import (
	"testing"
	"time"
)

func Test_Set(t *testing.T) {
	// Create a new instance of Datastore.
	datastore := NewDatastore()
	// Test case 1: set a value with expiry.
	key1 := "testkey1"
	value1 := "testvalue1"
	expiry1 := time.Duration(5)
	err := datastore.Set(key1, value1, expiry1)
	if err != nil {
		t.Errorf("Error setting value: %v", err)
	}

	// Test case 2: set a value without expiry.
	key2 := "testkey2"
	value2 := "testvalue2"
	err = datastore.Set(key2, value2, time.Duration(0))
	if err != nil {
		t.Errorf("Error setting value: %v", err)
	}

	// Test case 3: set a value with expiry of 0 seconds.
	key3 := "testkey3"
	value3 := "testvalue3"
	expiry3 := time.Duration(0)
	err = datastore.Set(key3, value3, expiry3)
	if err != nil {
		t.Errorf("Error setting value: %v", err)
	}

	// Test case 4: set a value with a very long expiry time.
	key4 := "testkey4"
	value4 := "testvalue4"
	expiry4 := time.Duration(99999999)
	err = datastore.Set(key4, value4, expiry4)
	if err != nil {
		t.Errorf("Error setting value: %v", err)
	}

	// Test case 5: set a value with the same key as an existing value.
	key5 := "testkey5"
	value5 := "testvalue5"
	expiry5 := time.Duration(5)
	err = datastore.Set(key5, value5, expiry5)
	if err != nil {
		t.Errorf("Error setting value: %v", err)
	}

	value5_new := "newtestvalue5"
	expiry5_new := time.Duration(10)
	err = datastore.Set(key5, value5_new, expiry5_new)
	if err != nil {
		t.Errorf("Error updating value: %v", err)
	}

	// Test case 6:check that the new value has been set.
	val, err := datastore.Get(key5)
	if err != nil {
		t.Errorf("Error getting value: %v", err)
	}
	if val != value5_new {
		t.Errorf("Incorrect value returned: got %v, expected %v", val, value5_new)
	}

	//Test case 7: test that item expires after 1 second
	datastore.Set("key1", "value1", time.Duration(1))
	time.Sleep(2 * time.Second)
	_, err = datastore.Get("key1")
	if err == nil {
		t.Errorf("Expected key1 to be expired but it was still found in datastore")
	}

	//Test case 8: test that item does not expire before specified time
	datastore.Set("key2", "value2", time.Duration(5))
	time.Sleep(2 * time.Second)
	_, err = datastore.Get("key2")
	if err != nil {
		t.Errorf("Expected key2 to still be present in datastore but got error: %v", err)
	}

	//Test case 9: test that item can be overwritten and still expire
	datastore.Set("key3", "value3", time.Duration(2))
	datastore.Set("key3", "newvalue", time.Duration(1))
	time.Sleep(2 * time.Second)
	_, err = datastore.Get("key3")
	if err == nil {
		t.Errorf("Expected key3 to be expired but it was still found in datastore")
	}
}
func Test_Get(t *testing.T) {
	// Initialize the data store
	datastore := NewDatastore()
	datastore.Set("key1", "value1", time.Duration(1))

	// Test 1: getting a valid key-value pair
	val, err := datastore.Get("key1")
	if err != nil {
		t.Errorf("Error while getting value for key1: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected value1, but got %v", val)
	}

	// Test 2: getting an expired key-value pair
	time.Sleep(time.Second * 2) // Wait for the key to expire
	_, err = datastore.Get("key1")
	if err == nil {
		t.Errorf("Expected ErrKeyNotFound, but got %v", err)
	}

	// Test 3: getting a non-existent key
	_, err = datastore.Get("non-existent-key")
	if err == nil {
		t.Errorf("Expected ErrKeyNotFound, but got %v", err)
	}
}
