package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {

	var cli bool
	var help bool
	var v bool
	flag.BoolVar(&v, "version", false, "Show version information")
	flag.BoolVar(&v, "v", false, "Show version information")
	flag.BoolVar(&cli, "cli", false, "Run in command-line interface mode")
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.BoolVar(&help, "h", false, "Show help message")
	flag.Parse()

	if v {
		commandExec([]string{"-v"})
		return
	} else if help {
		commandExec([]string{"-h"})
		return
	} else if cli {

		commandExec([]string{})

		/*
			fmt.Printf("**********%sGoManager%s**********\n", "\033[36m", "\033[0m")

			err := setupMasterPassword(false)
			if err != nil {
				fmt.Println("Error during setup:", err)
				return
			}

			tries := 0
			authenticated, err := AuthenticateUser()
			if err != nil {
				fmt.Println("Error during authentication:", err)
				return
			}

			for !authenticated && tries < 2 {
				fmt.Println("Invalid password, try again...")
				tries++
				authenticated, _ = AuthenticateUser()
			}
			if !authenticated {
				fmt.Println("Too many failed attempts. Exiting...")
				return
			}

			fmt.Println("Authentication successful!")

			for {

				fmt.Println("Choose an option: ")
				fmt.Println("1. Check saved passwords")
				fmt.Println("2. Save a new password")
				fmt.Println("3. Delete a saved password")
				fmt.Println("4. Search for a password")
				fmt.Println("5. Modify password")
				fmt.Println("6. Change master password")
				fmt.Println("7. Exit")

				var choice int
				fmt.Scanln(&choice)
				switch choice {
				case 1:
					clearScreen()
					fmt.Printf("%sWaring%s: Your passwords will be displayed in plain text.\n", "\033[31m", "\033[0m")
					tries := 0
					authenticated, err := AuthenticateUser()
					for !authenticated && tries < 2 {
						fmt.Println("Invalid password, try again...")
						tries++
						authenticated, _ = AuthenticateUser()
					}
					if !authenticated {
						fmt.Println("Too many failed attempts.")
						return
					}
					if err != nil {
						break
					}
					passwords, err := checkPasswords()
					if err != nil {
						fmt.Println("Error:", err)
					} else {
						fmt.Println("Saved Passowrds: ")
						for i := range passwords {
							fmt.Printf("\n--- Password %d ---\n", i+1)
							fmt.Printf("Service: %s\n", passwords[i].Service)
							fmt.Printf("Username: %s\n", passwords[i].Username)
							fmt.Printf("Password: %s\n", passwords[i].Password)
							fmt.Printf("Created: %s\n", passwords[i].Created)
							fmt.Print("----------------\n\n")
						}
					}
				case 2:
					clearScreen()
					passwords, err := checkPasswords()
					if err != nil && err.Error() != "no passwords saved" {
						fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						break
					}

					err = savePassword(passwords)
					if err != nil {
						fmt.Println("Error saving password:", err)
					} else {
						fmt.Println("Password saved successfully.")
					}
				case 3:
					clearScreen()
					passwords, err := checkPasswords()
					if err != nil && err.Error() != "no passwords saved" {
						fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						break
					}
					err = deletePassword(passwords)
					if err != nil {
						fmt.Println("Error deleting password:", err)
					} else {
						fmt.Println("Password deleted successfully.")
					}

				case 4:
					clearScreen()
					passwords, err := checkPasswords()
					if err != nil && err.Error() != "no passwords saved" {
						fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						break
					}
					found := searchPassword(passwords)

					if found != nil {
						tries := 0
						auth, err := AuthenticateUser()
						for !auth && tries < 2 {
							fmt.Println("Invalid password, try again...")
							tries++
							auth, _ = AuthenticateUser()
						}
						if !auth {
							fmt.Printf("Too many %sfailed%s attempts.\n", "\033[31m", "\033[0m")
							return
						}
						if err != nil {
							break
						}
						fmt.Println("Search Results: ")
						for i := range found {
							fmt.Printf("\n--- Password %d ---\n", i+1)
							fmt.Printf("Service: %s\n", found[i].Service)
							fmt.Printf("Username: %s\n", found[i].Username)
							fmt.Printf("Password: %s\n", found[i].Password)
							fmt.Printf("Created: %s\n", found[i].Created)
							fmt.Print("----------------\n\n")
						}
					}
				case 5:
					clearScreen()
					passwords, err := checkPasswords()
					if err != nil && err.Error() != "no passwords saved" {
						fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						break
					}
					err = modifyPassword(passwords)
					if err != nil {
						fmt.Println("Error modifying password:", err)
					} else {
						fmt.Println("Password modified successfully.")
					}

				case 6:
					clearScreen()
					err := changeMasterPassword()
					if err != nil {
						fmt.Printf("%sError changing master password:%s %s\n", "\033[31m", "\033[0m", err)
					} else {
						fmt.Println("Master password changed successfully.")
					}

				case 7:
					clearScreen()
					fmt.Println("Exiting...")
					return
				default:
					clearScreen()
					fmt.Println("Invalid choice. Please try again.")
				}
			}*/

	} else {

		if !HasDesktopEnvironment() {
			fmt.Printf("%sError%s: no desktop environment detected, cannot run GUI mode", "\033[31m", "\033[0m")
			exec.Command("")
		} else {
			hideConsole()
			gui()
		}
	}

}

func HasDesktopEnvironment() bool {
	switch runtime.GOOS {
	case "windows":
		return true
	case "linux":
		if os.Getenv("DISPLAY") != "" {
			return true
		}
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return true
		}
		if os.Getenv("DESKTOP_SESSION") != "" {
			return true
		}
		if os.Getenv("XDG_CURRENT_DESKTOP") != "" {
			return true
		}
		return false
	case "darwin":
		return true
	default:
		return false
	}
}

func commandExec(args []string) {

	filename := "./passport-cli"
	cmd := exec.Command(filename, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Errore durante l'avvio di %s: %v\n", filename, err)
		os.Exit(1)
	}
}
