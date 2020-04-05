package main

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/sid-sun/seaturtle"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	var toEncrypt bool
	var text, passPhrase []byte
	var outputPath string
	if len(os.Args) == 4 || len(os.Args) == 5 {
		if os.Args[1] == "-e" || os.Args[1] == "--encode" || os.Args[1] == "-encode" {
			toEncrypt = true
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "-help" {
			fmt.Printf("\nUsage:\n")
			fmt.Printf("    For encoding: %s (--encode / -encode / -e) <input file> <passphrase file> <output file (optional)>.\n", os.Args[0])
			fmt.Printf("    For decoding: %s (--decode / -decode / -d) <encrypted input> <passphrase file> <output file (optional)>.\n", os.Args[0])
		} else if !(os.Args[1] == "-d" || os.Args[1] == "--decode" || os.Args[1] == "-decode") {
			fmt.Println("Invalid argument:", os.Args[1])
		}
		if len(os.Args) == 5 {
			outputPath = os.Args[4]
		}
		text = readFromFile(os.Args[2])
		passPhrase = readFromFile(os.Args[3])
	} else {
		fmt.Printf("Usage:\n")
		fmt.Printf("    For encoding: %s (--encode / -encode / -e) <input file> <passphrase file> <output file (optional)>.\n", os.Args[0])
		fmt.Printf("    For decoding: %s (--decode / -decode / -d) <encrypted input> <passphrase file> <output file (optional)>.\n", os.Args[0])
		os.Exit(0)
	}
	if outputPath == "" {
		outputPath = os.Args[2] + ".seat"
	}
	key := sha256.Sum256(passPhrase)

	var output []byte
	var err error

	if toEncrypt {
		output, err = encrypt(key[:], text)
		if err != nil {
			panic(err.Error())
		}
	} else {
		output, err = decrypt(key[:], text)
		if err != nil {
			panic(err.Error())
		}
	}

	err = ioutil.WriteFile(outputPath, output, 0644)
	if err != nil {
		panic(err.Error())
	}
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := seaturtle.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, seaturtle.BlockSize+len(plaintext))

	iv := ciphertext[:seaturtle.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[seaturtle.BlockSize:], plaintext)

	return ciphertext, nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := seaturtle.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < seaturtle.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:seaturtle.BlockSize]
	ciphertext = ciphertext[seaturtle.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
