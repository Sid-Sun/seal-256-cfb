package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readFromFile(filePath string) []byte {
	// Check if file exists and if not, print
	if fileExists(filePath) {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err.Error())
		}
		return data
	}
	fmt.Println("File:", filePath, "seems to be nonexistent")
	os.Exit(0)
	return nil
}
