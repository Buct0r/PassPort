package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"golang.org/x/term"
)

const secret = "SECRET" // Can be changed to any desired file name
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

func setupMasterPassword(renew bool) error {

	path, err := getMasterPasswordPath()
	if err != nil {
		return err
	}

	if FileExists(path) && !renew {
		return nil
	}
	if !renew {
		if FileExists(filename) && filename != "" {
			os.Remove(filename)
		}
	}

	fmt.Printf("*** Welcome to %sGoManager CLI%s, the command line tool to securely store all of your credentials\n", "\033[36m", "\033[0m")
	fmt.Println("*** For the first start you will have to set a master password, that will allow you to access the application")
	fmt.Printf("*** Be %sVERY%s careful, as if you loose this password, you will no longer be able to access the other passwords\n", "\033[31m", "\033[0m")

	password, err := ReadPassword("Master Password: ")
	if err != nil {
		return err
	}

	for len(password) < 12 {
		fmt.Printf("%sError: Master password must be at least 12 characters%s\n", "\033[31m", "\033[0m")
		password, err = ReadPassword("Master Password: ")
		if err != nil {
			return err
		}
	}

	confirmation, err := ReadPassword("Confirm Master Password: ")
	if err != nil {
		return err
	}

	if !bytes.Equal(password, confirmation) {
		return fmt.Errorf("passwords do %snot%s match", "\033[31m", "\033[0m")
	} else {
		fmt.Printf("%sMaster password set successfully!%s\n\n", "\033[32m", "\033[0m")
	}

	for i := range confirmation {
		confirmation[i] = 0
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	//filename = rand.Text()

	return os.WriteFile(path, []byte(hashedPassword), 0600)

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

func savePassword(passwords []Password) error {

	newPassword := Password{}

	fmt.Printf("Enter %sservice%s: ", "\033[36m", "\033[0m")
	fmt.Scanln(&newPassword.Service)
	service, err := sanitizeInput([]byte(newPassword.Service))
	if err != nil {
		return err
	}
	newPassword.Service = string(service)
	fmt.Printf("Enter %susername%s: ", "\033[36m", "\033[0m")
	fmt.Scanln(&newPassword.Username)
	username, err := sanitizeInput([]byte(newPassword.Username))
	if err != nil {
		return err
	}
	newPassword.Username = string(username)
	password, err := ReadPassword("Enter password: ")
	if err != nil {
		return err
	}
	password, err = sanitizeInput(password)
	if err != nil {
		return err
	}
	newPassword.Password = string(password)
	date := time.Now().Format("2006-01-02 15:04:05") //TODO: give the possibilylity to change date format
	newPassword.Created = date

	passwords = append(passwords, newPassword)

	data, err := json.MarshalIndent(passwords, "", "  ")
	if err != nil {
		return err
	}

	/*
		encText, err := encryptData(filename, masterPasswordFile, string(data))
		if err != nil {
			return err
		}

		return os.WriteFile(filename, []byte(encText), 0644)
	*/

	path, err := getMasterPasswordPath()
	if err != nil {
		return err
	}

	text, err := encryptData(filename, path, string(data))
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(text), 0600)
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

func searchPassword(passwords []Password) []Password {
	if len(passwords) == 0 {
		fmt.Println("Error: No passwords saved")
		return nil
	}
	fmt.Println("Enter the service you are looking for: ")
	var service string
	fmt.Scanln(&service)

	var foundPasswords []Password

	for _, p := range passwords {
		if p.Service == service {
			foundPasswords = append(foundPasswords, p)
		}
	}

	if len(foundPasswords) == 0 {
		fmt.Println("Your search didn't generate any result")
		return nil
	}

	return foundPasswords

}

func deletePassword(passwords []Password) error {
	tries := 0

	authenticated, _ := AuthenticateUser()
	for !authenticated && tries < 2 {
		fmt.Println("Invalid password, try again...")
		tries++
		authenticated, _ = AuthenticateUser()
	}
	if !authenticated {
		fmt.Println("Too many failed attempts.")
		return nil
	}

	fmt.Println("Enter the service you want to delete: ")
	var service string
	fmt.Scanln(&service)

	var foundPasswords []Password

	for _, p := range passwords {
		if p.Service == service {
			foundPasswords = append(foundPasswords, p)
		}
	}

	if len(foundPasswords) == 0 {
		fmt.Println("Your search didn't generate any result")
		return nil
	}

	var passwordToDelete []Password
	var passwordToDeleteIndices []int

	for i, p := range passwords {
		if p.Service == service {
			passwordToDelete = append(passwordToDelete, p)
			passwordToDeleteIndices = append(passwordToDeleteIndices, i)
		}
	}

	fmt.Printf("\nFound %d entries for '%s'. Please choose which one to reveal:\n\n", len(foundPasswords), service)
	for i, p := range passwordToDelete {
		fmt.Printf("[%d] Username: %s | Created: %s\n", i+1, p.Username, p.Created)
	}

	var selectedIndex int
	if len(passwordToDelete) > 1 {
		fmt.Print("\nEnter your choice (1-" + fmt.Sprintf("%d", len(passwordToDelete)) + "): ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(passwordToDelete) {
			return fmt.Errorf("invalid choice")
		}
		selectedIndex = choice - 1
	} else {
		selectedIndex = 0
	}

	// Check if service was found
	if len(passwordToDelete) == 0 {
		return fmt.Errorf("service '%s' not found", service)
	}

	password := &passwordToDelete[selectedIndex]
	originalIndex := passwordToDeleteIndices[selectedIndex]
	// Show details before deletion
	fmt.Println("\n--- Password to be deleted ---")
	fmt.Printf("Service: %s\n", password.Service)
	fmt.Printf("Username: %s\n", password.Username)
	fmt.Printf("Created: %s\n", password.Created)
	fmt.Println("------------------------------")

	fmt.Println("\nAre you sure you want to delete this password? (yes/no): ")
	var choice string
	fmt.Scanln(&choice)

	if choice == "yes" {
		var updatedPasswords []Password
		for i, p := range passwords {
			if i != originalIndex {
				updatedPasswords = append(updatedPasswords, p)
			}
		}

		data, err := json.MarshalIndent(updatedPasswords, "", "  ")
		if err != nil {
			return err
		}
		path, err := getMasterPasswordPath()
		if err != nil {
			return err
		}
		text, err := encryptData(filename, path, string(data))
		if err != nil {
			return err
		}
		return os.WriteFile(filename, []byte(text), 0600)
	}

	return fmt.Errorf("deletion cancelled")
}

func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	case "linux", "darwin":
		cmd = exec.Command("clear")
	default:
		return
	}
	cmd.Stdout = os.Stdout

	_ = cmd.Run()

}

func changeMasterPassword() error {
	tries := 0
	authenticated, err := AuthenticateUser()
	if err != nil {
		fmt.Println("Error during authentication:")
		return err
	}

	for !authenticated && tries < 2 {
		fmt.Println("Invalid password, try again...")
		tries++
		authenticated, _ = AuthenticateUser()
	}
	if !authenticated {
		fmt.Println("Too many failed attempts. Exiting...")
		return nil
	}
	path, err := getMasterPasswordPath()
	if err != nil {
		return err
	}

	if FileExists(path) {
		data, error := decryptData(path, filename)

		fmt.Println("\n⚠️  You are about to change your master password")
		fmt.Println("Your current passwords will be re-encrypted with the new master password")
		fmt.Printf("Continue? (yes/no): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			fmt.Println("Operation cancelled. Your passwords are safe.")
			return nil
		}
		err := setupMasterPassword(true)
		if err != nil {
			fmt.Println("Error setting new master password. Your old master password is still valid.")
			return err
		}

		if !FileExists(filename) || error != nil {
			fmt.Printf("%sMaster password changed successfully!%s\n", "\033[32m", "\033[0m")
			return nil
		}

		text := []byte(data)
		var passwords []Password
		err = json.Unmarshal(text, &passwords)
		if err != nil {
			fmt.Printf("%sMaster password changed successfully!%s\n", "\033[32m", "\033[0m")
			return nil
		}

		if len(passwords) < 1 {
			fmt.Printf("%sMaster password changed successfully!%s\n", "\033[32m", "\033[0m")
			return nil
		}

		jsonData, err := json.MarshalIndent(passwords, "", "  ")
		if err != nil {
			return err
		}

		path, err := getMasterPasswordPath()
		if err != nil {
			return err
		}

		encText, err := encryptData(filename, path, string(jsonData))
		if err != nil {
			fmt.Println("Error re-encrypting passwords. Please try again.")
			return err
		}

		err = os.WriteFile(filename, []byte(encText), 0600)
		if err != nil {
			return err
		}

		fmt.Printf("%sMaster password changed successfully!%s\n", "\033[32m", "\033[0m")
		fmt.Printf("Your %d password(s) have been re-encrypted with the new master password\n", len(passwords))
		return nil

	} else {
		return fmt.Errorf("master password file does not exist")
	}
}

func sanitizeInput(input []byte) ([]byte, error) {
	input = bytes.TrimSpace(input)

	if len(input) == 0 {
		return nil, fmt.Errorf("input cannot be empty")
	}

	if len(input) > 255 {
		return nil, fmt.Errorf("input is too long")
	}

	for _, b := range input {
		if b < 32 || b == 127 {
			return nil, fmt.Errorf("invalid character in input")
		}
	}

	return input, nil
}

func modifyPassword(passwords []Password) error {
	fmt.Println("Enter the service you want to modify: ")
	var service string
	fmt.Scanln(&service)

	tries := 0

	authenticated, _ := AuthenticateUser()
	for !authenticated && tries < 2 {
		fmt.Println("Invalid password, try again...")
		tries++
		authenticated, _ = AuthenticateUser()
	}
	if !authenticated {
		fmt.Println("Too many failed attempts.")
		return nil
	}

	var matchingPasswords []Password
	var matchingIndices []int

	for i, p := range passwords {
		if p.Service == service {
			matchingPasswords = append(matchingPasswords, p)
			matchingIndices = append(matchingIndices, i)
		}
	}

	if len(matchingPasswords) == 0 {
		return fmt.Errorf("service '%s' not found", service)
	}

	var selectedPassword *Password
	var selectedIndex int

	if len(matchingPasswords) > 1 {
		fmt.Printf("\nFound %d entries for '%s'. Please choose which one to modify:\n\n", len(matchingPasswords), service)
		for i, p := range matchingPasswords {
			fmt.Printf("[%d] Username: %s | Created: %s\n", i+1, p.Username, p.Created)
		}
		fmt.Print("\nEnter your choice (1-" + fmt.Sprintf("%d", len(matchingPasswords)) + "): ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(matchingPasswords) {
			return fmt.Errorf("invalid choice")
		}
		selectedPassword = &matchingPasswords[choice-1]
		selectedIndex = matchingIndices[choice-1]
	} else {
		selectedPassword = &matchingPasswords[0]
		selectedIndex = matchingIndices[0]
	}

	fmt.Println("Enter new details (leave blank to keep current value):")

	fmt.Printf("New service (current: %s): ", selectedPassword.Service)
	var newService string
	fmt.Scanln(&newService)
	if newService != "" {
		sanitizedService, err := sanitizeInput([]byte(newService))
		if err != nil {
			return err
		}
		selectedPassword.Service = string(sanitizedService)
	}

	fmt.Printf("New username (current: %s): ", selectedPassword.Username)
	var newUsername string
	fmt.Scanln(&newUsername)
	if newUsername != "" {
		sanitizedUsername, err := sanitizeInput([]byte(newUsername))
		if err != nil {
			return err
		}
		selectedPassword.Username = string(sanitizedUsername)
	}
	fmt.Printf("New password (leave blank to keep current): ")
	newPassword, err := ReadPassword("")
	if err != nil {
		return err
	}
	if len(newPassword) > 0 {
		sanitizedPassword, err := sanitizeInput(newPassword)
		if err != nil {
			return err
		}
		selectedPassword.Password = string(sanitizedPassword)
	}

	passwords[selectedIndex] = *selectedPassword

	data, err := json.MarshalIndent(passwords, "", "  ")
	if err != nil {
		return err
	}
	path, err := getMasterPasswordPath()
	if err != nil {
		return err
	}
	text, err := encryptData(filename, path, string(data))
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(text), 0600)

}

func maskPassword(password string) string {
	masked := ""
	for range password {
		masked += "*"
	}
	return masked
}

func revealPassword(passwords []Password) error {
	fmt.Println("Enter the service you want to reveal the password for: ")
	var service string
	fmt.Scanln(&service)

	var foundPasswords []Password

	for _, p := range passwords {
		if p.Service == service {
			foundPasswords = append(foundPasswords, p)
		}
	}

	if len(foundPasswords) == 0 {
		fmt.Println("Your search didn't generate any result")
		return nil
	}

	fmt.Printf("\nFound %d entries for '%s'. Please choose which one to reveal:\n\n", len(foundPasswords), service)
	for i, p := range foundPasswords {
		fmt.Printf("[%d] Username: %s | Created: %s\n", i+1, p.Username, p.Created)
	}
	if len(foundPasswords) != 1 {
		fmt.Print("\nEnter your choice (1-" + fmt.Sprintf("%d", len(foundPasswords)) + "): ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(foundPasswords) {
			return fmt.Errorf("invalid choice")
		}
		selectedPassword := foundPasswords[choice-1]

		fmt.Printf("\n%sWarning:%s Revealing passwords can be risky. Make sure no one is looking at your screen.\n", "\033[31m", "\033[0m")
		time.Sleep(2 * time.Second)

		fmt.Printf("Password for service %s'%s'%s and username %s'%s'%s : %s\n", "\033[32m", selectedPassword.Service, "\033[0m", "\033[32m", selectedPassword.Username, "\033[0m", selectedPassword.Password)
		fmt.Printf("%sThis password will be hidden again in 10 seconds%s\n", "\033[31m", "\033[0m")
		time.Sleep(10 * time.Second)
		clearScreen()

		return nil
	} else {
		fmt.Printf("\n%sWarning:%s Revealing passwords can be risky. Make sure no one is looking at your screen.\n", "\033[31m", "\033[0m")
		time.Sleep(2 * time.Second)
		fmt.Printf("Password for service %s'%s'%s and username %s'%s'%s : %s\n", "\033[32m", foundPasswords[0].Service, "\033[0m", "\033[32m", foundPasswords[0].Username, "\033[0m", foundPasswords[0].Password)
		fmt.Printf("%sThis password will be hidden again in 10 seconds%s\n", "\033[31m", "\033[0m")
		time.Sleep(10 * time.Second)
		clearScreen()
		return nil
	}
}
