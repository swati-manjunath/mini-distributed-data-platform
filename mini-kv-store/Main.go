package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("invalid %s value %q: %v", key, value, err))
	}

	return n
}

func main() {
	// Load local overrides when a .env file exists. Defaults keep a fresh clone runnable.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	// Command-line flags
	port := flag.Int("port", 8080, "Port to run server on")
	nodeID := flag.Int("node-id", 1, "Unique node ID")
	clusterFlag := flag.String(
		"cluster",
		"1=127.0.0.1:8080,2=127.0.0.1:8081,3=127.0.0.1:8082",
		"Cluster configuration",
	)

	flag.Parse()

	// Parse cluster configuration
	cluster = parseCluster(*clusterFlag, *nodeID)
	numberOfNodes = len(cluster.Nodes)

	configuredNodeCount := getEnvInt("NUMBER_OF_NODES", numberOfNodes)
	if configuredNodeCount != numberOfNodes {
		panic(fmt.Sprintf(
			"NUMBER_OF_NODES=%d does not match cluster node count=%d",
			configuredNodeCount,
			numberOfNodes,
		))
	}

	fmt.Printf("Node %d starting at %s\n", cluster.Self.ID, cluster.Self.Address)
	fmt.Println("Known cluster nodes:")
	for _, n := range cluster.Nodes {
		fmt.Printf("  Node %d -> %s\n", n.ID, n.Address)
	}

	// Register handlers
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/put", handlePostRequest)
	http.HandleFunc("/get", handleGetRequest)
	http.HandleFunc("/replicate", handleReplicateRequest)
	http.HandleFunc("/history", handleHistoryRequest)
	http.HandleFunc("/latest", handleLatestRequest)

	// WAL file for the current node. Each line is a JSON-encoded PutRequest.
	dataFileName = fmt.Sprintf("data-%d.log", cluster.Self.ID)

	// Load persisted key/value data
	loadDataFromFile()

	// Start HTTP server
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Listening on %s...\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
