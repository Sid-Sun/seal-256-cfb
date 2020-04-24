package main

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/sid-sun/sealion"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	var toEncrypt bool
	var text, passPhrase []byte
	var outputPath string
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
		text = readFromFile(os.Args[2])
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
	block, err := sealion.NewCipher(key)
	if err != nil {
		return nil, err
	}

	originalPlaintextLength := len(plaintext)

	emptyBytes := make([]byte, sealion.BlockSize)
	plaintext = append(plaintext, emptyBytes...)

	iv := plaintext[originalPlaintextLength:]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(plaintext[:originalPlaintextLength], plaintext[:originalPlaintextLength])

	//Rotate plaintext by the original text's length so that the iv bytes are in the front again
	plaintext = append(plaintext[originalPlaintextLength:], plaintext[:originalPlaintextLength]...)

	return plaintext, nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := sealion.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < sealion.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:sealion.BlockSize]
	ciphertext = ciphertext[sealion.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
