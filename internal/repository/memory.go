package repository

import (
	"errors"
	"fmt"
	"sync"

	"kvstore/internal/wal"
)

type KVMap map[string]string

type InMemory struct {
	mut *sync.RWMutex 
	wal *wal.WAL
	appendToWALFlag bool
	kvmap KVMap
}

func NewInMemoryStore(w *wal.WAL)*InMemory{
	// initialize the map (and mutex) to prevent
	// panic: runtime error: assignment to entry in nil map
	return &InMemory{
		mut: new(sync.RWMutex),
		wal: w,
		kvmap: make(map[string]string),
	}
}

func (store *InMemory) EnableWAL() {
	store.appendToWALFlag = true
}

func (store *InMemory) GetCurrentState() KVMap {
	return store.kvmap
} 

func (store *InMemory) SetState(lastState map[string]string) error {
	for key, value := range lastState {
		err := store.Set(key, value)
		if err != nil {
			return fmt.Errorf("'%v' on SET operation with key : %s", err, key)
		}
	}
	return nil
}

func (store *InMemory) Apply(records []wal.Record) error {
	// iterate over records
	for index, record := range records{
		// match the operation
		switch record.Operation{
		case wal.Set:
			key := record.Key
			value := record.Value
			err := store.Set(key, value)
			if err != nil {
				return fmt.Errorf("'%v' on SET operation with index(0-based) %d", err, index)
			}
		case wal.Update:
			key := record.Key
			value := record.Value
			err := store.Update(key, value)
			if err != nil {
				return fmt.Errorf("'%v' on UPDATE operation with index(0-based) %d", err, index)
			}
		case wal.Delete:
			key := record.Key
			err := store.Delete(key)
			if err != nil {
				return fmt.Errorf("'%v' on DELETE operation with index(0-based) %d", err, index)
			}
		case wal.Clear:
			err := store.Clear()
			if err != nil {
				return fmt.Errorf("'%v' on CLEAR operation with index(0-based) %d", err, index)
			}
		} 
	}
	return nil
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

	// first log the operation to wal
	if store.appendToWALFlag {
		record := store.wal.Serialize("SET", key, value)
		err := store.wal.Write(record)
		if err != nil {
			return err
		}
	}

	// set the key 
	store.kvmap[key] = value
	return nil
}

func (store *InMemory)Get(key string) (string, error) {
	store.mut.RLock()
	defer store.mut.RUnlock()

	exists := store.keyExists(key)
	if !exists {
		return "", errors.New("key not found")
	}

	value := store.kvmap[key]
	return value, nil
}

func (store *InMemory)Update(key string, value string) error {
	store.mut.Lock()
	defer store.mut.Unlock()

	exists := store.keyExists(key)
	if !exists{
		return errors.New("key not found")
	}
	
	// first log the operation to wal
	if store.appendToWALFlag {
		record := store.wal.Serialize("UPDATE", key, value)
		err := store.wal.Write(record)
		if err != nil {
			return err
		}
	}

	store.kvmap[key] = value
	return nil
}

func (store *InMemory)Delete(key string) error {
	store.mut.Lock()
	defer store.mut.Unlock()

	exists := store.keyExists(key)
	if !exists{
		return errors.New("key not found")
	}

	// first log the operation to wal
	if store.appendToWALFlag {
		record := store.wal.Serialize("DELETE", key, "")
		err := store.wal.Write(record)
		if err != nil {
			return err
		}
	}

	delete(store.kvmap, key)
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
	for key := range store.kvmap{
		allkeys = append(allkeys, key)
	}

	return allkeys
}

func (store *InMemory) Clear() error {
	/*
	TWO CHOICES
	1 either clear the wal.log file so that old operations are removed hence space is saved
	2 or keep the clear operation as it is and skip the entries before clear while replaying
	
	if we clear the wal.log file and then the program crashes before clearing the in-memory
	then there is no way to recover the latest state
	Hence, i'll keep the wal.log file as it is for the MVP
	*/
	store.mut.Lock()
	defer store.mut.Unlock()

	// first log the operation to wal
	if store.appendToWALFlag {
		record := store.wal.Serialize("CLEAR", "", "")
		err := store.wal.Write(record)
		if err != nil {
			return err
		}
	}

	clear(store.kvmap)
	return nil
}

/* === HELPERS === */

func (store *InMemory)keyExists(key string) bool {
	// assume that the caller already does mut.Lock i.e. prevent deadlock

	_, ok := store.kvmap[key]
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
	for k,v := range store.kvmap{
		store.Set(k, v)
	}
}