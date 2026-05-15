package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	var req PutRequest
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value cannot be empty", http.StatusBadRequest)
		return
	}
	writeIntoFile(bodyBytes)
	mu.Lock()
	fmt.Printf("Received PUT request for key: %s, value: %s\n", req.Key, req.Value)
	store[req.Key] = req.Value
	mu.Unlock()
	fmt.Fprintf(w, "Key-value pair stored!")
	fmt.Printf("Current store: %v\n", store)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mu.RLock()
	defer mu.RUnlock()
	key := r.URL.Query().Get("key")
	value, exists := store[key]
	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Value for key '%s': %s", key, value)
}
