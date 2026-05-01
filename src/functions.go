package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/pbkdf2"

	"golang.org/x/term"
)

// Can be changed to any desired file name
const secret = "SECRET"
const masterPasswordFile = "master.key"

var filename string

func init() {
	// Initialize filename at startup
	path, err := getSecretFilePath()
	if err != nil {
		// Fall back to a temporary location if path resolution fails
		filename = filepath.Join(os.TempDir(), "passport_secret")
	} else {
		filename = path
	}
}

type Password struct {
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
	Created  string `json:"created"` //date
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func ReadPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return password, err
}

func HashPassword(password []byte) (string, error) {

	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := pbkdf2.Key(password, salt, 100000, 32, sha256.New)

	result := append(salt, hash...)

	for i := range password {
		password[i] = 0
	}
	return hex.EncodeToString(result), nil
}

func VerifyPassword(password []byte, storedHashHex string) (bool, error) {
	storedHashBytes, err := hex.DecodeString(storedHashHex)
	if err != nil {
		return false, fmt.Errorf("authentication failed")
	}

	if len(storedHashBytes) < 16 {
		return false, fmt.Errorf("authentication failed")
	}

	salt := storedHashBytes[:16]
	storedHash := storedHashBytes[16:]

	computedHash := pbkdf2.Key(password, salt, 100000, 32, sha256.New)

	for i := range password {
		password[i] = 0
	}

	return subtle.ConstantTimeCompare(computedHash, storedHash) == 1, nil
}

func AuthenticateUser() (bool, error) {
	path, err := getMasterPasswordPath()
	if err != nil {
		return false, fmt.Errorf("authentication failed")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("authentication failed")
	}

	storedHash := string(data)
	password, err := ReadPassword("Enter Master Password: ")
	if err != nil {
		return false, fmt.Errorf("authentication failed")
	}

	match, err := VerifyPassword(password, storedHash)
	if err != nil {
		return false, fmt.Errorf("authentication failed")
	}

	if !match {
		return false, fmt.Errorf("authentication failed")
	}

	return true, nil
}

func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no passwords saved")
	}
	return lines, nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func checkPasswords() ([]Password, error) {
	if !FileExists(filename) {
		_, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		return []Password{}, fmt.Errorf("no passwords saved")
	}

	path, err := getMasterPasswordPath()
	if err != nil {
		return nil, err
	}

	data, err := decryptData(path, filename)

	/*
		if err.Error() != "ciphertext too short" {
			return nil, err
		}
	*/

	text := []byte(data)
	if err != nil {
		return nil, err
	}
	var passwords []Password
	err = json.Unmarshal(text, &passwords)
	if len(passwords) < 1 {
		return nil, fmt.Errorf("no passwords saved")
	}
	return passwords, err
}
