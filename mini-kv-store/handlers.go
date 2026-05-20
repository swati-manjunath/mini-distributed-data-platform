package main

import (
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
	defer r.Body.Close()

	// Read the request body once so it can be reused if forwarding.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse JSON.
	var req PutRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value cannot be empty", http.StatusBadRequest)
		return
	}

	// Forward if this node is not the owner.
	targetNodeID := hashRing.getNodeForKey(req.Key)
	if !isLocalNode(targetNodeID) {
		fmt.Printf("Forwarding post request for key %q to node %d\n", req.Key, targetNodeID)
		forwardPostRequest(w, targetNodeID, bodyBytes, req)
		return
	}

	// This node owns the key: persist and store locally.
	writeIntoFile(bodyBytes)

	putInStore(req.Key, req.Value)
	fmt.Printf("Stored key=%q value=%q on node %d\n",
		req.Key,
		req.Value,
		cluster.Self.ID,
	)
	fmt.Printf("Current store: %v\n", store)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetNodeID := hashRing.getNodeForKey(r.URL.Query().Get("key"))

	if !isLocalNode(targetNodeID) {
		fmt.Printf("Forwarding get request for key %q to node %d\n", r.URL.Query().Get("key"), targetNodeID)
		forwardGetRequest(w, targetNodeID, r.URL.Query().Get("key"))
		return
	}

	value, exists := getFromStore(r.URL.Query().Get("key"))
	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Value for key '%s': %s", r.URL.Query().Get("key"), value)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"node":   cluster.Self.Address,
	})
}
