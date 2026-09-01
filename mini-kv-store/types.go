package main

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Node struct {
	ID      int
	Address string // e.g. "127.0.0.1:8080"
}

type Config struct {
	Self  Node
	Nodes []Node
}

type HashRing struct {
	keys         []uint32
	nodes        map[uint32]Node
	virtualNodes int
}
