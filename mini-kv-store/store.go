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
	if !ok || len(keys) == 0 {
		mu.RUnlock()
		return nil, false
	}

	keysCopy := append([]string(nil), keys...)
	history := make([]string, 0, len(keys))
	for _, k := range keysCopy {
		v := store[k]
		history = append(history, v)
	}
	mu.RUnlock()
	return history, true
}

func getLatestFromStore(key string) (string, bool) {
	mu.RLock()
	keys, ok := index[key]
	if !ok || len(keys) == 0 {
		mu.RUnlock()
		return "", false
	}
	latest := keys[len(keys)-1]
	value, exists := store[latest]
	mu.RUnlock()
	return value, exists
}
