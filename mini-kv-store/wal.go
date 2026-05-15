package main

import (
	"fmt"
	"os"
)

func writeIntoFile(bodyBytes []byte) {
	file, err := os.OpenFile("data.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
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
