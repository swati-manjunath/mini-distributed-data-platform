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

	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value cannot be empty", http.StatusBadRequest)
		return
	}

	hostName := getHostNameFromKey(req.Key)
	targetNodeID := hashRing.getNodeForKey(hostName)

	if !isLocalNode(targetNodeID) {
		forwardPostRequest(w, targetNodeID, bodyBytes, req)
		return
	}

	writeIntoFile(bodyBytes)

	putInStore(req.Key, req.Value, hostName)

	replicaNode := getReplicaNode(targetNodeID)
	go sendReplicationRequest(bodyBytes, req, replicaNode)

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

	shouldReturn := tryReplicaRead(exists, replicaNode, w, "/get", requestKey)
	if shouldReturn {
		return
	}

	if exists {
		fmt.Fprintf(w, "Value for key '%s': %s", requestKey, value)
	} else {
		http.Error(w, "Key not found", http.StatusNotFound)

	}
}

func tryReplicaRead(exists bool, replicaNode Node, w http.ResponseWriter, path string, requestKey string) bool {
	// Current failover logic assumes a single replica per primary.
	// Multi-hop forwarding would require loop-prevention metadata.
	if !exists && replicaNode.ID != cluster.Self.ID {
		fmt.Printf("Forwarding request to replica node %d for read\n", replicaNode.ID)
		forwardGetPathRequest(w, replicaNode.ID, path, requestKey)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func getHostNameFromKey(key string) string {
	return strings.Split(key, SEPARATOR)[0]
}
