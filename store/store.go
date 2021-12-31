package store

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"
)

const fileName = "retain.db"

type Storage struct {
	internal sync.Map
}

// New will return a new instance of store.Storage
// It will have content loaded from disk if retain.db exists.
func New() *Storage {

	storage := Storage{
		internal: *new(sync.Map),
	}

	loadedFromDisk := storage.LoadFromDisk(fileName)
	if loadedFromDisk {
		fmt.Printf("Loaded from disk: %s", fileName)
	}

	return &storage
}

// Get gives you the value stored at key
func (storage *Storage) Get(key string) (interface{}, bool) {

	value, ok := storage.internal.Load(key)
	if !ok {
		return nil, false
	}
	return value, ok
}

// Set lets you store/update a key-value pair
func (storage *Storage) Set(key string, value interface{}) {

	storage.internal.Store(key, value)
}

// Delete will wipe out the relevant key-value pair
func (storage *Storage) Delete(key string) {

	storage.internal.Delete(key)
}

// LoadFromDisk lets you load your data from a given path
func (storage *Storage) LoadFromDisk(path string) bool {

	file, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	handleError("LoadFromDisk: file open", err)
	defer file.Close()

	var temp map[string]interface{}
	err = gob.NewDecoder(file).Decode(&temp)
	handleError("LoadFromDisk: file decode", err)

	storage.internal = *toInternalMap(&temp)
	return true
}

// Save will dump the in-memory map to disk
func (storage *Storage) Save() {

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	handleError("Save: file open", err)
	defer file.Close()

	temp := fromInternalMap(&storage.internal)

	err = gob.NewEncoder(file).Encode(temp)
	handleError("Save: file encode", err)
}

func toInternalMap(temp *map[string]interface{}) *sync.Map {

	m := sync.Map{}
	for key, value := range *temp {
		m.Store(key, value)
	}
	return &m
}

func fromInternalMap(m *sync.Map) *map[string]interface{} {

	temp := make(map[string]interface{})
	m.Range(func(key interface{}, value interface{}) bool {
		temp[key.(string)] = value
		return true
	})
	return &temp
}

func handleError(text string, err error) {

	if err != nil {
		log.Fatal(text, err)
	}
}
