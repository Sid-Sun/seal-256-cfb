package main

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sid-sun/sealion"
)

func main() {
	var toEncrypt bool
	var outputPath string
	var err error

	var sealionCipher cipher.Block
	var passPhrase []byte

	var inputStream, outputStream chan []byte
	var wg sync.WaitGroup

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

		passPhrase = readFromFile(os.Args[3])
		key := sha256.Sum256(passPhrase)

		sealionCipher, err = sealion.NewCipher(key[:])
		if err != nil {
			panic(err.Error())
		}

		samples := 10
		sampleBytes := make([]byte, sealion.BlockSize)
		var avg time.Duration

		for i := 0; i < samples; i++ {
			t0 := time.Now()
			sealionCipher.Encrypt(sampleBytes, sampleBytes)
			avg += time.Now().Sub(t0)
		}
		rate := (int64(sealion.BlockSize*samples) / avg.Microseconds()) * sealion.BlockSize * 1000

		inputStream = make(chan []byte, rate) // 65536 - 1048576 - 524288 - 540672 - 655360*
		outputStream = make(chan []byte, rate)

		wg.Add(1)
		go startReader(os.Args[2], &inputStream, &wg)

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
	// key := sha256.Sum256(passPhrase)
	// outputStream := make(chan []byte, )

	wg.Add(2)
	if toEncrypt {
		// go encrypt(key[:], &inputStream, &outputStream, &wg)
		go encrypt(&sealionCipher, &inputStream, &outputStream, &wg)
	} else {
		go decrypt(&sealionCipher, &inputStream, &outputStream, &wg)
	}
	go startWriter(outputPath, &outputStream, &wg)

	wg.Wait()
}

func encrypt(sealionCipher *cipher.Block, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	// sealionCipher, err := sealion.NewCipher(key)
	// if err != nil {
	// 	panic(err.Error())
	// }

	iv := make([]byte, (*sealionCipher).BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err.Error())
	}

	//copy(iv, "Welll, this good")

	// Push IV to output stream
	*outputStream <- iv

	// Create new CFB Encryptor
	cfb := cipher.NewCFBEncrypter(*sealionCipher, iv)

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

func decrypt(sealionCipher *cipher.Block, inputStream, outputStream *chan []byte, wg *sync.WaitGroup) {
	// seaLionCipher, err := sealion.NewCipher(key)
	// if err != nil {
	// 	panic(err.Error())
	// }

	iv := <-*inputStream
	cfb := cipher.NewCFBDecrypter(*sealionCipher, iv)

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
