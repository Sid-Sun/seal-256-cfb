package main

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sid-sun/sealion"
	"golang.org/x/crypto/sha3"
)

const version string = "1.0.0" // Program Version

func main() {
	var toEncrypt bool
	var outputPath string
	var err error

	var blockCipher cipher.Block
	var inputStream, outputStream chan []byte
	var progressStream chan int64
	var wg sync.WaitGroup

	if len(os.Args) == 4 || len(os.Args) == 5 {
		switch os.Args[1] {
		case "-e", "--encrypt":
			toEncrypt = true
		case "-h", "--help":
			printHelp()
		case "-d", "--decrypt":
		default:
			fmt.Println("Invalid argument:", os.Args[1])
			os.Exit(1)
		}

		// If the input file does not exist, print so and quit
		if !fileExists(os.Args[2]) {
			fmt.Println("File:", os.Args[2], "seems to be nonexistent")
			os.Exit(1)
		}

		// If there's a 5th distinct arguement, treat it as output file name
		if len(os.Args) == 5 && os.Args[4] != os.Args[2] {
			outputPath = os.Args[4]
		}

		// Read contents of passphrase file and pass it through SHA-256
		key := sha3.Sum256(readFromFile(os.Args[3]))

		// Create the sealion cipher from the hash of passphrase
		blockCipher, err = sealion.NewCipher(key[:])
		if err != nil {
			panic(err.Error())
		}

		// Run 10 samples through standard cipher block encryption
		// And measure the average time taken to encrypt a block
		// So we can make a buffer of the appropriate size for the current machine
		samples := 10
		sampleBytes := make([]byte, blockCipher.BlockSize())
		var avg time.Duration

		for i := 0; i < samples; i++ {
			t0 := time.Now()
			blockCipher.Encrypt(sampleBytes, sampleBytes)
			avg += time.Now().Sub(t0)
		}

		// Calculate the buffer size by finding the average number of bytes encrypted
		// Each nano second (in float) and multiplying the result by the blocksize and 1000 * 1000
		// The multiplication by these specific magic numbers is based on some logic which I came up with
		// But failed to properly document - will be explained at some point in the unforseeable future
		bufferSize := int64(float64(blockCipher.BlockSize()*samples) / float64(avg.Nanoseconds()) * float64(blockCipher.BlockSize()) * 1000000)

		// Create buffere input, output and progress channels with the calculated buffer size
		inputStream = make(chan []byte, bufferSize)
		outputStream = make(chan []byte, bufferSize)
		progressStream = make(chan int64, bufferSize)

		// Start the reader
		wg.Add(1)
		go readInput(os.Args[2], blockCipher.BlockSize(), &inputStream, &progressStream, &wg)
	} else {
		// Check for version
		if len(os.Args) > 1 {
			if os.Args[1] == "-v" || os.Args[1] == "--version" {
				fmt.Println(version)
				os.Exit(0)
			}
		}
		printHelp()
		os.Exit(1)
	}

	// If the ouput file was not specified append .seal to the input file
	// And use it as output file path
	if outputPath == "" {
		outputPath = os.Args[2] + ".seal"
	}

	// Start the encrypt / decrypt and output writer routines
	wg.Add(2)
	if toEncrypt {
		go encrypt(&blockCipher, &inputStream, &outputStream, &wg)
	} else {
		go decrypt(&blockCipher, &inputStream, &outputStream, &wg)
	}
	go writeOutput(outputPath, &outputStream, &wg)

	// The first output from progress stream is total file size
	// Send that as target to progressbar function along with the progress stream
	// For subsequent progress reporting
	progressBar(<-progressStream, &progressStream)

	// Wait until all goroutines have ended before quitting
	wg.Wait()
}

func encrypt(blockCipher *cipher.Block, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	// Defer waitgroup go-routine done before returning
	defer wg.Done()

	// Generate a random initialisation vector for CFB
	iv := make([]byte, (*blockCipher).BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err.Error())
	}

	// Push IV to output stream
	*outputStream <- iv

	// Create CFB Encrypter
	cfb := cipher.NewCFBEncrypter(*blockCipher, iv)

	for {
		// Fetch full / part block from inputStream
		block := <-*inputStream

		// A nil in block signals end of file
		if block == nil {
			// Push nil to output stream and break out of loop
			*outputStream <- nil
			break
		}

		// Run CFB on block
		cfb.XORKeyStream(block, block)
		// Push block to output stream
		*outputStream <- block
	}
}

func decrypt(blockCipher *cipher.Block, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	// Defer waitgroup go-routine done before returning
	defer wg.Done()

	// First input from input stream is the IV
	iv := <-*inputStream

	// Create CFB decrypter with IV
	// Potential error of IV not being of the same size as cipher Block Size
	// Is handled in NewCFBDecrypter already
	cfb := cipher.NewCFBDecrypter(*blockCipher, iv)

	for {
		// Read block from input stream
		block := <-*inputStream

		// A nil block signals end of input
		if block == nil {
			// Push nil to output stream and break out of loop
			*outputStream <- nil
			break
		}

		// Run CFB on block
		cfb.XORKeyStream(block, block)
		// Push block to output stream
		*outputStream <- block
	}
}
