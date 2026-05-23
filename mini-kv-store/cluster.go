package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
)

var cluster Config // Global cluster config (accessible from handlers if needed)
var numberOfNodes int
var hashRing HashRing

// Parse "1=127.0.0.1:8080,2=127.0.0.1:8081,3=127.0.0.1:8082"
func parseCluster(clusterStr string, selfID int) Config {
	var nodes []Node
	var self Node
	found := false

	entries := strings.Split(clusterStr, ",")

	for _, entry := range entries {
		parts := strings.Split(entry, "=")
		if len(parts) != 2 {
			panic("invalid cluster entry: " + entry)
		}

		var id int
		_, err := fmt.Sscanf(parts[0], "%d", &id)
		if err != nil {
			panic("Invalid node ID in cluster entry: " + entry)
		}

		node := Node{
			ID:      id,
			Address: parts[1],
		}

		nodes = append(nodes, node)

		if id == selfID {
			self = node
			found = true
		}
	}

	if !found {
		panic(fmt.Sprintf("Self node ID %d not found in cluster config", selfID))
	}

	return Config{
		Self:  self,
		Nodes: nodes,
	}
}

func (h HashRing) getNodeForKey(key string) int {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))
	hash := hasher.Sum32()

	fmt.Printf("Key '%s' hashed to %d\n", key, hash)
	return int((hash % uint32(numberOfNodes)) + 1) // Node IDs are 1-based
}

func isLocalNode(node int) bool {
	return node == cluster.Self.ID
}

func forwardPostRequest(w http.ResponseWriter, targetNodeID int, bodyBytes []byte, req PutRequest) {
	var targetAddress string
	found := false

	for _, node := range cluster.Nodes {
		if node.ID == targetNodeID {
			targetAddress = node.Address
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Target node not found in cluster config", http.StatusInternalServerError)
		return
	}

	fmt.Printf(
		"Forwarding post request for key %q to node %d at %s\n",
		req.Key,
		targetNodeID,
		targetAddress,
	)

	resp, err := http.Post(
		fmt.Sprintf("http://%s/put", targetAddress),
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}

	// Return success to the original client.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"forwarded"}`))
}

func forwardGetRequest(w http.ResponseWriter, targetNodeID int, key string) {
	var targetAddress string
	found := false

	for _, node := range cluster.Nodes {
		if node.ID == targetNodeID {
			targetAddress = node.Address
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Target node not found in cluster config", http.StatusInternalServerError)
		return
	}

	fmt.Printf(
		"Forwarding get request for key %q to node %d at %s\n",
		key,
		targetNodeID,
		targetAddress,
	)
	resp, err := http.Get(
		fmt.Sprintf("http://%s/get?key=%s", targetAddress, key),
	)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
	responseBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func getReplicaNode(primaryNodeID int) Node {
	for _, node := range cluster.Nodes {
		if node.ID == primaryNodeID {
			if node.ID == numberOfNodes {
				return cluster.Nodes[0]
			}
			return cluster.Nodes[node.ID]
		}
	}
	panic("Primary node ID not found in cluster config")
}

func sendReplicationRequest(w http.ResponseWriter, bodyBytes []byte, req PutRequest, targetNode Node) {
	fmt.Printf(
		"Forwarding key %q to node %d at %s for replication\n",
		req.Key,
		targetNode.ID,
		targetNode.Address,
	)

	resp, err := http.Post(
		fmt.Sprintf("http://%s/replicate", targetNode.Address),
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}

	// Return success to the original client.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"forwarded"}`))
}
