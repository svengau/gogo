package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
)

// get AES Gaulois/Counter Mode
// https://en.wikipedia.org/wiki/Galois/Counter_Mode
func getAESGCM(keyString string) cipher.AEAD {
	//key, err := hex.DecodeString(keyString)
	// if err != nil {
	// 	log.Fatal("An error occured while decoding password. Error: ", err)
	// }
	key := []byte(keyString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("An error occured while creating cipher with key %v. Error: ", keyString, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal("An error occured while creating GCM. Error: ", err)
	}
	return aesGCM
}

func encryptString(stringToEncrypt string, keyString string) (encryptedString string) {
	aesGCM := getAESGCM(keyString)

	nonce := make([]byte, aesGCM.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal("An error occured while creating nonce. Error: ", err)
	}

	plaintext := []byte(stringToEncrypt)
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func decryptString(encryptedString string, keyString string) (decryptedString string) {
	aesGCM := getAESGCM(keyString)

	nonceSize := aesGCM.NonceSize()
	enc, _ := hex.DecodeString(encryptedString)
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}
