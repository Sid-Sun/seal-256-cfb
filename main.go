package main

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sid-sun/sealion"
)

func main() {
	var toEncrypt bool
	var passPhrase []byte
	var outputPath string

	var wg sync.WaitGroup
	inputStream := make(chan []byte, 65536)

	if len(os.Args) == 4 || len(os.Args) == 5 {
		if os.Args[1] == "-e" || os.Args[1] == "--encrypt" || os.Args[1] == "-encrypt" {
			toEncrypt = true
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "-help" {
			fmt.Printf("\nUsage:\n")
			fmt.Printf("    For encryption: %s (--encrypt / -encrypt / -e) <input file> <passphrase file> <output file (optional)>.\n", os.Args[0])
			fmt.Printf("    For decryption: %s (--decrypt / -decrypt / -d) <encrypted input> <passphrase file> <output file (optional)>.\n", os.Args[0])
		} else if !(os.Args[1] == "-d" || os.Args[1] == "--decrypt" || os.Args[1] == "-decrypt") {
			fmt.Println("Invalid argument:", os.Args[1])
		}

		if len(os.Args) == 5 {
			outputPath = os.Args[4]
		}

		wg.Add(1)
		go startReader(os.Args[2], &inputStream, &wg)

		passPhrase = readFromFile(os.Args[3])
	} else {
		fmt.Printf("Usage:\n")
		fmt.Printf("    For encryption: %s (--encrypt / -encrypt / -e) <input file> <passphrase file> <output file (optional)>.\n", os.Args[0])
		fmt.Printf("    For decryption: %s (--decrypt / -decrypt / -d) <encrypted input> <passphrase file> <output file (optional)>.\n", os.Args[0])
		os.Exit(0)
	}

	if outputPath == "" {
		outputPath = os.Args[2] + ".seal"
	}

	// Run passphrase through SHA256 to get key
	key := sha256.Sum256(passPhrase)
	outputStream := make(chan []byte, 65536)

	wg.Add(2)
	if toEncrypt {
		go encrypt(key[:], &inputStream, &outputStream, &wg)
	} else {
		go decrypt(key[:], &inputStream, &outputStream, &wg)
	}
	go startWriter(outputPath, &outputStream, &wg)

	wg.Wait()
}

func encrypt(key []byte, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	seaLionCipher, err := sealion.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	iv := make([]byte, seaLionCipher.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err.Error())
	}

	//copy(iv, "Welll, this good")

	// Push IV to output stream
	*outputStream <- iv

	// Create new CFB Encryptor
	cfb := cipher.NewCFBEncrypter(seaLionCipher, iv)

	for {
		block := <-*inputStream
		if block == nil {
			// Push nil to output stream and break out of loop
			*outputStream <- nil
			break
		}

		// Run CFB on blockbyte
		cfb.XORKeyStream(block, block)
		// Push block to output stream
		*outputStream <- block
	}

	wg.Done()
}

func decrypt(key []byte, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	seaLionCipher, err := sealion.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	iv := <-*inputStream
	cfb := cipher.NewCFBDecrypter(seaLionCipher, iv)

	for {
		block := <-*inputStream
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

	wg.Done()
}
