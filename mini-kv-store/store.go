package main

import (
	"sync"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)
