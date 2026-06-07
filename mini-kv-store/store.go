package main

import (
	"sync"
)

var (
	store = make(map[string]string)
	index = make(map[string][]string)
	mu    sync.RWMutex
)

func putInStore(key, value string, hostName string) {
	mu.Lock()
	store[key] = value
	index[hostName] = append(index[hostName], key)
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

func getHistoryFromStore(key string) ([]string, bool) {
	mu.RLock()
	keys, ok := index[key]
	mu.RUnlock()
	if !ok || len(keys) == 0 {
		return nil, false
	}

	history := make([]string, 0, len(keys))
	for _, k := range keys {
		mu.RLock()
		v := store[k]
		mu.RUnlock()
		history = append(history, v)
	}
	return history, true
}

func getLatestFromStore(key string) (string, bool) {
	mu.RLock()
	keys, ok := index[key]
	if !ok || len(keys) == 0 {
		return "", false
	}
	latest := keys[len(keys)-1]
	mu.RUnlock()
	mu.RLock()
	value, exists := store[latest]
	mu.RUnlock()
	return value, exists
}
