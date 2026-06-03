package main

import (
	"fmt"
	"sync"
)

var (
	store = make(map[string]string)
	index = make(map[string][]string) // For storing history of values for each key
	mu    sync.RWMutex
)

func putInStore(key, value string, hostName string) {
	mu.Lock()
	store[key] = value
	// Store the key in the history index for the host so we can lookup values
	index[hostName] = append(index[hostName], key) // Append key to history
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
	fmt.Printf("Index %v\n", index)
	fmt.Printf("key %s\n", key)
	keys, ok := index[key]
	if !ok || len(keys) == 0 {
		fmt.Printf("No keys found for host %s\n", key)
		return "", false
	}
	latest := keys[len(keys)-1] // Get the most recent key for this host
	fmt.Printf("Latest key for host %s is %s\n", key, latest)
	mu.RUnlock()
	mu.RLock()
	value, exists := store[latest]
	mu.RUnlock()
	return value, exists
}
