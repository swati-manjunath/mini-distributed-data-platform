package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func loadDataFromFile() {
	file, err := os.Open("data.log")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No existing data file found. Starting with an empty store.")
			return
		}
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var req PutRequest
		err := json.Unmarshal([]byte(line), &req)
		if err != nil {
			fmt.Printf("Skipping invalid line in data.log: %s (error: %s)\n", line, err.Error())
			continue
		}
		store[req.Key] = req.Value
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %s\n", err.Error())
	}
}

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {

	loadDataFromFile()

	http.HandleFunc("/put", handlePostRequest)

	http.HandleFunc("/get", handleGetRequest)

	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
