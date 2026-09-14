package repository

import (
	"errors"
	"kvstore/internal/wal"
	"sync"
)

type InMemory struct {
	mut *sync.RWMutex 
	wal *wal.WAL
	kvpair map[string]string
}

func NewInMemoryStore(w *wal.WAL)*InMemory{
	// initialize the map (and mutex) to prevent
	// panic: runtime error: assignment to entry in nil map
	return &InMemory{
		mut: new(sync.RWMutex),
		wal: w,
		kvpair: make(map[string]string),
	}
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

var SeedData = map[string]string{
	"name" : "yingshai",
	"game" : "nine sols",
	"x" : "12",
	"pi" : "3.14", 
}
func (store *InMemory)Seed(){
	for k,v := range store.kvpair{
		store.Set(k, v)
	}
}