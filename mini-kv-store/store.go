package main

import (
	"sync"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func putInStore(key, value string) {
	mu.Lock()
	store[key] = value
	mu.Unlock()

}
func getFromStore(key string) (string, bool) {
	mu.RLock()
	value, exists := store[key]
	mu.RUnlock()
	return value, exists
}

func deleteFromStore(key string) {
	mu.Lock()
	delete(store, key)
	mu.Unlock()
}
