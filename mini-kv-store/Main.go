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
	// Check if the environment variable is set and parse it as an integer. If not set or invalid, return the default value.
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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
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

	numberOfNodes = getEnvInt("NUMBER_OF_NODES", 3)
	fmt.Printf("Loaded NUMBER_OF_NODES=%d from .env\n", numberOfNodes)

	// Parse cluster configuration
	cluster = parseCluster(*clusterFlag, *nodeID)

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
