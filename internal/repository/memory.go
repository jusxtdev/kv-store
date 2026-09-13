package repository

import (
	"errors"
	"sync"
)

type InMemory struct {
	mut sync.RWMutex 
	kvpair map[string]string
}

func (store *InMemory)Set(key string, value string) error {
	// mutex lock
	store.mut.Lock()
	defer store.mut.Unlock()

	// check if key exists
	exists := store.keyExists(key)
	if exists {
		return errors.New("key already exists")
	}

	// set the key 
	store.kvpair[key] = value
	return nil
}

func (store *InMemory)Get(key string) (string, error) {
	store.mut.RLock()
	defer store.mut.RUnlock()

	exists := store.keyExists(key)
	if !exists {
		return "", errors.New("key not found")
	}

	value := store.kvpair[key]
	return value, nil
}

func (store *InMemory)Update(key string, value string) error {
	store.mut.Lock()
	defer store.mut.Unlock()

	exists := store.keyExists(key)
	if !exists{
		return errors.New("key not found")
	}

	store.kvpair[key] = value
	return nil
}

func (store *InMemory)Delete(key string) error {
	store.mut.Lock()
	defer store.mut.Unlock()

	exists := store.keyExists(key)
	if !exists{
		return errors.New("key not found")
	}

	delete(store.kvpair, key)
	return nil
}

func (store *InMemory)Exists(key string) bool {
	store.mut.RLock()
	defer store.mut.RUnlock()

	exists := store.keyExists(key)
	if exists{
		return true
	}
	return false
}

func (store *InMemory)Keys() []string {
	store.mut.RLock()
	defer store.mut.RUnlock()

	var allkeys []string
	for key := range store.kvpair{
		allkeys = append(allkeys, key)
	}

	return allkeys
}

func (store *InMemory)Clear() {
	store.mut.Lock()
	defer store.mut.Unlock()

	clear(store.kvpair)
}

/* === HELPERS === */

func (store *InMemory)keyExists(key string) bool {
	// assume that the caller already does mut.Lock i.e. prevent deadlock

	_, ok := store.kvpair[key]
	if ok {
		return true
	} else {
		return false
	}
}