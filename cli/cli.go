package main

import (
	"flag"
	"fmt"
	"time"
)

const version = "0.4.0" //TODO: UPDATE EVERY TIME A CHANGE IS MADE TO THE CLI

func main() {

	var cli bool
	var help bool
	var v bool
	flag.BoolVar(&v, "version", false, "Show version information")
	flag.BoolVar(&v, "v", false, "Show version information")
	flag.BoolVar(&cli, "cli", false, "Run in command-line interface mode")
	flag.BoolVar(&cli, "c", false, "Run in command-line interface mode")
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.BoolVar(&help, "h", false, "Show help message")
	flag.Parse()

	if v {
		fmt.Printf("PassPort version %s\n", version)
		return
	} else if help {
		fmt.Println("PassPort CLI Help")
		fmt.Println("Usage: ")
		fmt.Println("  -version, -v Show version information")
		fmt.Println("  -cli, -c        Run in command-line interface mode")
		fmt.Println("  -help, -h   Show this help message")
		return
	} else {

		fmt.Printf("**********%sPassPort%s**********\n", "\033[36m", "\033[0m")

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

		start := time.Now()
		const sessionTimeout = 60 * time.Second
	Loop:

		for {
			if time.Since(start) > sessionTimeout {
				fmt.Printf("%sSession expired due to inactivity.%s\n", "\033[31m", "\033[0m")
				break
			}

			fmt.Println("Choose an option: ")
			fmt.Println("1. Check saved passwords")
			fmt.Println("2. Save a new password")
			fmt.Println("3. Delete a saved password")
			fmt.Println("4. Search for a password")
			fmt.Println("5. Modify password")
			fmt.Println("6. Change master password")
			fmt.Println("7. Exit")
			var choice int
			if _, err := fmt.Scanln(&choice); err != nil {
				var discard string
				fmt.Scanln(&discard)
				fmt.Println("Invalid input. Please enter a number.")
				continue
			}

			switch choice {
			case 1:
				clearScreen()
				//fmt.Printf("%sWaring%s: Your passwords will be displayed in plain text.\n", "\033[31m", "\033[0m")

				passwords, err := checkPasswords()
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Saved Passowrds: ")
					for i := range passwords {
						fmt.Printf("\n--- Password %d ---\n", i+1)
						fmt.Printf("Service: %s\n", passwords[i].Service)
						fmt.Printf("Username: %s\n", passwords[i].Username)
						fmt.Printf("Password: %s\n", maskPassword((passwords[i].Password)))
						fmt.Printf("Created: %s\n", passwords[i].Created)
						fmt.Print("----------------\n\n")
					}

					fmt.Println("(1) Reveal Password")
					fmt.Println("(2) Back to Main Menu")
					var choice int
					fmt.Scanln(&choice)
					switch choice {
					case 1:
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
						err = revealPassword(passwords)
						if err != nil {
							fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						}
					case 2:
						continue
					default:
						fmt.Println("Invalid choice. Returning to main menu.")
						continue

					}
				}
				start = time.Now() // Reset session timer after activity
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
				start = time.Now() // Reset session timer after activity
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
				start = time.Now() // Reset session timer after activity
			case 4:
				clearScreen()
				passwords, err := checkPasswords()
				if err != nil && err.Error() != "no passwords saved" {
					fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
					break Loop
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
						break Loop
					}
					fmt.Println("Search Results: ")
					for i := range found {
						fmt.Printf("\n--- Password %d ---\n", i+1)
						fmt.Printf("Service: %s\n", found[i].Service)
						fmt.Printf("Username: %s\n", found[i].Username)
						fmt.Printf("Password: %s\n", maskPassword(found[i].Password))
						fmt.Printf("Created: %s\n", found[i].Created)
						fmt.Print("----------------\n\n")
					}
					fmt.Println("(1) Reveal Password")
					fmt.Println("(2) Back to Main Menu")
					var choice int
					fmt.Scanln(&choice)
					switch choice {
					case 1:
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
						err = revealPassword(passwords)
						if err != nil {
							fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
						}
					case 2:
						continue
					default:
						fmt.Println("Invalid choice. Returning to main menu.")
						continue

					}

				}
				start = time.Now() // Reset session timer after activity
			case 5:
				clearScreen()
				passwords, err := checkPasswords()
				if err != nil && err.Error() != "no passwords saved" {
					fmt.Printf("%sError:%s %s\n", "\033[31m", "\033[0m", err)
					break Loop
				}
				err = modifyPassword(passwords)
				if err != nil {
					fmt.Println("Error modifying password:", err)
				} else {
					fmt.Println("Password modified successfully.")
				}
				start = time.Now() // Reset session timer after activity
			case 6:
				clearScreen()
				err := changeMasterPassword()
				if err != nil {
					fmt.Printf("%sError changing master password:%s %s\n", "\033[31m", "\033[0m", err)
				} else {
					fmt.Println("Master password changed successfully.")
				}
				start = time.Now() // Reset session timer after activity
			case 7:
				clearScreen()
				fmt.Println("Exiting...")
				start = time.Now() // Reset session timer after activity
				return
			default:
				clearScreen()
				start = time.Now() // Reset session timer after activity
				fmt.Println("Invalid choice. Please try again.")
			}

		}
	}

}
