#!/bin/bash

# Build the binary
go build -o server .

echo "Starting servers..."

# Full cluster configuration (same for every node)
CLUSTER="1=127.0.0.1:8080,2=127.0.0.1:8081,3=127.0.0.1:8082"

# Start each node with:
# -node-id : identifies this node
# -port    : port to listen on
# -cluster : full list of all nodes
./server -node-id=1 -port=8080 -cluster="$CLUSTER" > server_8080.log 2>&1 &
./server -node-id=2 -port=8081 -cluster="$CLUSTER" > server_8081.log 2>&1 &
./server -node-id=3 -port=8082 -cluster="$CLUSTER" > server_8082.log 2>&1 &

echo "Servers are running."
echo "Cluster: $CLUSTER"
echo "Logs: server_8080.log, server_8081.log, server_8082.log"

wait