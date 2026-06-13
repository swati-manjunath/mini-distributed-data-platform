package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
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

	nodeIndex := int(hash % uint32(len(cluster.Nodes)))
	return cluster.Nodes[nodeIndex].ID
}

func isLocalNode(node int) bool {
	return node == cluster.Self.ID
}

func findNodeAddress(targetNodeID int) (string, bool) {
	for _, node := range cluster.Nodes {
		if node.ID == targetNodeID {
			return node.Address, true
		}
	}

	return "", false
}

func copyForwardedResponse(w http.ResponseWriter, resp *http.Response) {
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Failed to copy forwarded response: %v\n", err)
	}
}

func forwardPostRequest(w http.ResponseWriter, targetNodeID int, bodyBytes []byte, req PutRequest) {
	targetAddress, found := findNodeAddress(targetNodeID)
	if !found {
		http.Error(w, fmt.Sprintf("Target node %d not found in cluster config", targetNodeID), http.StatusBadGateway)
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
		http.Error(w, "Failed to forward post request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyForwardedResponse(w, resp)
}

func forwardGetRequest(w http.ResponseWriter, targetNodeID int, key string) {
	forwardGetPathRequest(w, targetNodeID, "/get", key)
}

func forwardGetPathRequest(w http.ResponseWriter, targetNodeID int, path string, key string) {
	targetAddress, found := findNodeAddress(targetNodeID)
	if !found {
		http.Error(w, fmt.Sprintf("Target node %d not found in cluster config", targetNodeID), http.StatusBadGateway)
		return
	}

	fmt.Printf(
		"Forwarding %s request for key %q to node %d at %s\n",
		path,
		key,
		targetNodeID,
		targetAddress,
	)
	resp, err := http.Get(
		fmt.Sprintf("http://%s%s?key=%s", targetAddress, path, url.QueryEscape(key)),
	)
	if err != nil {
		fmt.Printf("Failed to forward %s request to node %d: %v\n", path, targetNodeID, err)
		http.Error(w, "Failed to forward get request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyForwardedResponse(w, resp)
}

func getReplicaNode(primaryNodeID int) Node {
	for i, node := range cluster.Nodes {
		if node.ID == primaryNodeID {
			replicaIndex := (i + 1) % len(cluster.Nodes)
			return cluster.Nodes[replicaIndex]
		}
	}
	panic("Primary node ID not found in cluster config")
}

func sendReplicationRequest(bodyBytes []byte, req PutRequest, targetNode Node) {
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

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Replication request to node %d failed with %d: %s\n", targetNode.ID, resp.StatusCode, string(responseBody))
		return
	}
}

func forwardGetHistoryRequest(w http.ResponseWriter, targetNodeID int, key string) {
	forwardGetPathRequest(w, targetNodeID, "/history", key)
}

func forwardGetLatestRequest(w http.ResponseWriter, targetNodeID int, key string) {
	forwardGetPathRequest(w, targetNodeID, "/latest", key)
}
