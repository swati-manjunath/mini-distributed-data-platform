package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

var dataFileName string

func writeIntoFile(bodyBytes []byte) {
	file, err := os.OpenFile(dataFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bodyString := string(bodyBytes)

	_, err = file.WriteString(bodyString + "\n")
	if err != nil {
		fmt.Printf("Error writing to file: %s\n", err.Error())
	}
}

func loadDataFromFile() {
	file, err := os.Open(dataFileName)
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
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			fmt.Printf("Skipping invalid line in data.log: %s (error: %v)\n", line, err)
			continue
		}

		store[req.Key] = req.Value
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}
