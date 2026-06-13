package main

import (
	"encoding/json"
	"net/http"
)

type historyResponse struct {
	Key     string   `json:"key"`
	History []string `json:"history"`
}

type latestResponse struct {
	Key    string `json:"key"`
	Latest string `json:"latest"`
}

func handleHistoryRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestKey := r.URL.Query().Get("key")
	targetNodeID := hashRing.getNodeForKey(requestKey)
	replicaNode := getReplicaNode(targetNodeID)

	if !isLocalNode(targetNodeID) && cluster.Self.ID != replicaNode.ID {
		forwardGetHistoryRequest(w, targetNodeID, requestKey)
		return
	}

	value, exists := getHistoryFromStore(requestKey)

	shouldReturn := tryReplicaRead(exists, replicaNode, w, "/history", requestKey)

	if shouldReturn {
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(historyResponse{Key: requestKey, History: value})
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key not found"})

	}
}

func handleLatestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestKey := r.URL.Query().Get("key")
	targetNodeID := hashRing.getNodeForKey(requestKey)
	replicaNode := getReplicaNode(targetNodeID)

	if !isLocalNode(targetNodeID) && cluster.Self.ID != replicaNode.ID {
		forwardGetLatestRequest(w, targetNodeID, requestKey)
		return
	}

	value, exists := getLatestFromStore(requestKey)

	shouldReturn := tryReplicaRead(exists, replicaNode, w, "/latest", requestKey)

	if shouldReturn {
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(latestResponse{Key: requestKey, Latest: value})
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key not found"})

	}
}
