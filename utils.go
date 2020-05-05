package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/sid-sun/sealion"
)

func readInput(fileName string, stream *chan []byte, progressStream *chan int64, wg *sync.WaitGroup) {
	// Defer waitgroup go-routine done before returning
	defer wg.Done()

	// Open input file
	file, err := os.Open(fileName)
	if err != nil {
		panic(err.Error())
	}

	// Defer file close and panic if there is an error
	defer func() {
		if err := file.Close(); err != nil {
			panic(err.Error())
		}
	}()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err.Error())
	}

	// Get input file size
	fileSize := fileInfo.Size()

	// Push input file size as the first input to progress stream
	*progressStream <- fileSize

	// Initialize offset
	offset := int64(0)
	for {
		if fileSize-offset >= sealion.BlockSize {
			// Create full-block
			block := make([]byte, sealion.BlockSize)

			// Read from file at offset to full-block
			// Panic if there are any errors
			bytesRead, err := file.ReadAt(block, offset)
			if err != nil {
				panic(err.Error())
			}

			// PUSH full-block to buffered s channel
			*stream <- block
			// Increment offset by bytesRead
			offset += int64(bytesRead)

			// Push offset to buffered progress channel
			*progressStream <- offset
		} else if fileSize-offset == 0 {
			// Exit condition - the entire file is read, exit
			// Send nil to stream to signal end of input
			// Break out of loop
			*stream <- nil
			break
		} else {
			// Create partial block
			block := make([]byte, fileSize-offset)

			// Read from file at offset to part-block
			// Panic if there are any errors
			bytesRead, err := file.ReadAt(block, offset)
			if err != nil {
				panic(err.Error())
			}

			// PUSH partial-block to channel
			*stream <- block
			// Increment offset by bytesRead
			offset += int64(bytesRead)

			// Push offset to buffered progress channel
			*progressStream <- offset
		}
	}
}

func writeOutput(fileName string, stream *chan []byte, wg *sync.WaitGroup) {
	// Defer waitgroup go-routine done before returning
	defer wg.Done()

	// Create output file
	file, err := os.Create(fileName)
	if err != nil {
		panic(err.Error())
	}

	// Defer file close and panic if there is an error
	defer func() {
		if err := file.Close(); err != nil {
			panic(err.Error())
		}
	}()

	// Initialize offset
	offset := int64(0)
	for {
		// Read block from stream
		block := <-*stream

		// A nil block signals end, break out of loop
		if block == nil {
			// I WANT TO BREAK FREE
			// FROM FOR LOOP
			break
		}

		// Write block to file at offset
		// Panic if there are any errors
		bytesWritten, err := file.WriteAt(block, offset)
		if err != nil {
			panic(err.Error())
		}

		// Increment offset by bytesWritten
		offset += int64(bytesWritten)
	}
}

func progressBar(fileSize int64, progressStream *chan int64) {
	// Create new progressbar with count
	bar := pb.StartNew(int(fileSize))
	for offset := 0; offset < int(fileSize); {
		offset = int(<-*progressStream)
		// Set bar to current offset from reader
		bar.Set(offset)
	}
	bar.Finish()
}

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
	// If file doesn't exist, print so and exit
	fmt.Println("File:", filePath, "seems to be nonexistent")
	os.Exit(1)
	// Fun fact: if you remove the below return
	// Compile will fail - as functions which return must
	// return at the end, BUT... The above OS Exit makes it
	// such that it'll never execute XD
	return nil
}

func printHelp() {
	// Ah, yes; help.
	fmt.Printf("%s is a CLI program which implements the SeaLion Block Cipher (http://github.com/sid-sun/sealion) in CFB (cipher feedback) mode with 256-Bit key length, using SHA3-256.", os.Args[0])
	fmt.Printf("\nDeveloped by Sidharth Soni (Sid Sun) <sid@sidsun.com>")
	fmt.Printf("\nOpen-sourced under The Unlicense")
	fmt.Printf("\nSource Code: http://github.com/sid-sun/seal-256-cfb\n")
	fmt.Printf("\nUsage:\n")
	fmt.Printf("    To encrypt: %s (--encrypt / -e) <input file> <passphrase file> <output file (optional)>\n", os.Args[0])
	fmt.Printf("    To decrypt: %s (--decrypt / -d) <encrypted input> <passphrase file> <output file (optional)>\n", os.Args[0])
	fmt.Printf("    To get version number: %s (--version / -v)\n", os.Args[0])
	fmt.Printf("    To get help: %s (--help / -h)\n", os.Args[0])
}
