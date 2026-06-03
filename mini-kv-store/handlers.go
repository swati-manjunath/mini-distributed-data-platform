package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var SEPARATOR string = "_"

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

	hostName := getHostNameFromKey(req.Key)
	// Forward if this node is not the owner.
	targetNodeID := hashRing.getNodeForKey(hostName)
	if !isLocalNode(targetNodeID) {
		forwardPostRequest(w, targetNodeID, bodyBytes, req)
		return
	}

	// This node owns the key: persist and store locally.
	writeIntoFile(bodyBytes)

	putInStore(req.Key, req.Value, hostName)
	fmt.Printf("Stored key=%q value=%q on node %d\n",
		req.Key,
		req.Value,
		cluster.Self.ID,
	)

	// Send replication request to other nodes
	replicaNode := getReplicaNode(targetNodeID)
	sendReplicationRequest(w, bodyBytes, req, replicaNode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestKey := r.URL.Query().Get("key")
	targetNodeID := hashRing.getNodeForKey(requestKey)
	replicaNode := getReplicaNode(targetNodeID)

	if !isLocalNode(targetNodeID) && cluster.Self.ID != replicaNode.ID {
		forwardGetRequest(w, targetNodeID, requestKey)
		return
	}

	value, exists := getFromStore(requestKey)

	shouldReturn := tryReplicaRead(exists, replicaNode, w, requestKey)
	if shouldReturn {
		return
	}

	if exists {
		fmt.Fprintf(w, "Value for key '%s': %s", requestKey, value)
	} else {
		http.Error(w, "Key not found", http.StatusNotFound)

	}
}

func tryReplicaRead(exists bool, replicaNode Node, w http.ResponseWriter, requestKey string) bool {
	// Current failover logic assumes a single replica per primary.
	// Multi-hop forwarding would require loop-prevention metadata.
	if !exists && replicaNode.ID != cluster.Self.ID {
		fmt.Printf("Forwarding request to replica node %d for read\n", replicaNode.ID)
		forwardGetRequest(w, replicaNode.ID, requestKey)
		return true
	}
	return false
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"node":   cluster.Self.Address,
	})
}

func handleReplicateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	var req PutRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	hostName := getHostNameFromKey(req.Key)
	writeIntoFile(bodyBytes)
	putInStore(req.Key, req.Value, hostName)

	fmt.Printf("Replicated key=%q value=%q on node %d\n",
		req.Key,
		req.Value,
		cluster.Self.ID,
	)
	fmt.Printf("Current store: %v\n", store)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"replicated"}`))
}

func getHostNameFromKey(key string) string {
	return strings.Split(key, SEPARATOR)[0]
}
