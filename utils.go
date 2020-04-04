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
	if fileExists(filePath) {
		data, err := ioutil.ReadFile(filePath)
		// if our program was unable to read the file
		// print out the reason why it can't
		if err != nil {
			panic(err.Error())
		}
		return data
		// if it was successful in reading the file then
		// print out the contents as a string
	} else {
		fmt.Println("File:", filePath, "seems to be nonexistent")
		os.Exit(0)
		return nil
	}
}
