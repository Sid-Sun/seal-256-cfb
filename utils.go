package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/sid-sun/sealion"
)

func startReader(fileName string, stream *chan []byte, progressStream *chan int64, wg *sync.WaitGroup) {
	if fileExists(fileName) {
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

		fileInfo, err := file.Stat()
		if err != nil {
			panic(err.Error())
		}

		fileSize := fileInfo.Size()

		*progressStream <- fileSize

		offset := int64(0)
		for {
			if fileSize-offset >= sealion.BlockSize {
				block := make([]byte, sealion.BlockSize)
				_, err = file.ReadAt(block, offset)
				if err != nil {
					panic(err.Error())
				}

				// PUSH full-block to channel
				*stream <- block
				offset += sealion.BlockSize
				*progressStream <- offset
			} else if fileSize-offset == 0 { // Once the entire file is read, exit
				// Send nil to stream to signal end
				*stream <- nil
				break
			} else {
				block := make([]byte, fileSize-offset)
				_, err := file.ReadAt(block, offset)
				if err != nil {
					panic(err.Error())
				}

				// PUSH partial-block to channel
				*stream <- block
				offset += int64(len(block))
				*progressStream <- offset
			}
		}
	} else {
		panic("Input File does not exist")
	}

	wg.Done()
}

func progressBar(count int64, progressStream *chan int64) {
	bar := pb.StartNew(int(count))
	var prev int
	var i int
	for i = 0; i < int(count); {
		i = int(<-*progressStream)
		bar.Add(i - prev)
		prev = i
	}
	bar.Finish()
}

func startWriter(fileName string, stream *chan []byte, wg *sync.WaitGroup) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err.Error())
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err.Error())
		}
	}()

	offset := int64(0)
	for {
		block := <-*stream
		if block == nil {
			break
		}
		_, err := file.WriteAt(block, offset)
		if err != nil {
			panic(err.Error())
		}

		offset += int64(len(block))
	}

	wg.Done()
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
	fmt.Println("File:", filePath, "seems to be nonexistent")
	os.Exit(0)
	return nil
}
