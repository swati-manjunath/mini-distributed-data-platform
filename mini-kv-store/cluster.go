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
		fmt.Printf("Target node %d not found in cluster config\n", targetNodeID)
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
		fmt.Printf("Failed to forward post request to node %d: %v\n", targetNodeID, err)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
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
		fmt.Printf("Target node %d not found in cluster config\n", targetNodeID)
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
		fmt.Printf("Failed to forward get request to node %d: %v\n", targetNodeID, err)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
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
		fmt.Printf("Failed to forward replication request to node %d: %v\n", targetNode.ID, err)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
}

func forwardGetHistoryRequest(w http.ResponseWriter, targetNodeID int, key string) {
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
		fmt.Printf("Target node %d not found in cluster config\n", targetNodeID)
		return
	}

	fmt.Printf(
		"Forwarding get history request for key %q to node %d at %s\n",
		key,
		targetNodeID,
		targetAddress,
	)
	resp, err := http.Get(
		fmt.Sprintf("http://%s/history?key=%s", targetAddress, key),
	)
	if err != nil {
		fmt.Printf("Failed to forward get history request to node %d: %v\n", targetNodeID, err)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
}

func forwardGetLatestRequest(w http.ResponseWriter, targetNodeID int, key string) {
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
		fmt.Printf("Target node %d not found in cluster config\n", targetNodeID)
	}

	fmt.Printf(
		"Forwarding get latest request for key %q to node %d at %s\n",
		key,
		targetNodeID,
		targetAddress,
	)
	resp, err := http.Get(
		fmt.Sprintf("http://%s/latest?key=%s", targetAddress, key),
	)
	if err != nil {
		fmt.Printf("Failed to forward get latest request to node %d: %v\n", targetNodeID, err)
		return
	}
	defer resp.Body.Close()

	// Propagate any error from the destination node.
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(responseBody), resp.StatusCode)
		return
	}
}
