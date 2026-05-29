package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func encryptData(filename string, key string, text string) (string, error) {

	if !FileExists(key) {
		return "", fmt.Errorf("key file does not exist")
	}

	chKeyHex, err := os.ReadFile(key)
	if err != nil {
		return "", err
	}

	defer zeroBytes(chKeyHex)

	chKey, err := hex.DecodeString(string(chKeyHex))

	defer zeroBytes(chKey)

	if err != nil {
		return "", fmt.Errorf("invalid key format")
	}

	if len(chKey) != 48 {
		return "", fmt.Errorf("invalid key size: expected 48 bytes from master.key, got %d", len(chKey))
	}

	aesKey := make([]byte, 32)
	copy(aesKey, chKey[16:48])
	defer zeroBytes(aesKey)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(text), nil)

	return hex.EncodeToString(ciphertext), nil

}

func decryptData(key string, passowrds string) (string, error) {
	if !FileExists(key) {
		return "", fmt.Errorf("key file does not exist")
	}

	chKeyHex, err := os.ReadFile(key)
	if err != nil {
		return "", err
	}

	defer zeroBytes(chKeyHex)

	chKey, err := hex.DecodeString(string(chKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid key format")
	}
	defer zeroBytes(chKey)
	// master.key contains: salt (16 bytes) + hash (32 bytes) = 48 bytes total
	if len(chKey) != 48 {
		return "", fmt.Errorf("invalid key size: expected 48 bytes from master.key, got %d", len(chKey))
	}
	// Use only the hash part (last 32 bytes) as AES key
	aesKey := make([]byte, 32)
	copy(aesKey, chKey[16:48])
	defer zeroBytes(aesKey)

	/*
		chKey, err := os.ReadFile(key)
		if err != nil {
			return "", err
		}

		chKey = chKey[:32]
	*/

	encText, err := os.ReadFile(passowrds)
	if err != nil {
		return "", err
	}

	if len(encText) == 0 {
		return "[]", nil
	}

	ciphertext, err := hex.DecodeString(string(encText))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:nonceSize]
	actualCiphertext := ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", fmt.Errorf("authentication failed: file may be corrupted or tampered with - %v", err)
	}
	return string(plaintext), nil
}
