//go:build windows || linux
// +build windows linux

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func getCardBackground(themeName string) *canvas.Rectangle {
	switch themeName {
	case "DarkTheme":
		return canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	case "LightTheme":
		return canvas.NewRectangle(color.RGBA{R: 220, G: 220, B: 220, A: 255})
	case "CustomTheme":
		return canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	default:
		return canvas.NewRectangle(color.RGBA{R: 20, G: 40, B: 60, A: 255})
	}
}

var guiLimiter = NewRateLimiter(5, 60*time.Second) // Allow 5 attempts per minute for GUI

func authWindow(a fyne.App, onComplete func(bool)) {
	w := a.NewWindow("PassPort: Authentication")
	w.Resize(fyne.NewSize(500, 800))

	// Flag per tracciare se il callback è stato già eseguito
	completed := false

	// Gestisci la chiusura manuale della finestra
	w.SetOnClosed(func() {
		if !completed {
			onComplete(false)
		}
	})

	path, err := getMasterPasswordPath()
	if err != nil {
		dialog.ShowError(fmt.Errorf("error determining master password path: %v", err), w)
		completed = true
		onComplete(false)
		return
	}

	img := canvas.NewImageFromResource(resourceLogoPng)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(500, 500))

	if FileExists(path) {
		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Enter Master Password")

		submitButton := widget.NewButton("Submit", func() {
			// Check if rate limiter is locked
			if locked, remaining := guiLimiter.IsLocked(); locked {
				seconds := int(remaining.Seconds())
				dialog.ShowError(
					fmt.Errorf("Too many failed attempts. Try again in %d second(s)", seconds),
					w,
				)
				return
			}

			if authenticate(passwordEntry.Text) {
				guiLimiter.RecordSuccess()
				completed = true
				w.Close()
				onComplete(true)
			} else {
				guiLimiter.RecordFailure()
				remainingAttempts := guiLimiter.GetRemainingAttempts()

				if locked, remaining := guiLimiter.IsLocked(); locked {
					seconds := int(remaining.Seconds())
					dialog.ShowInformation("Error", fmt.Sprintf("Too many failed attempts. Locked for %d second(s).", seconds), w)
					time.AfterFunc(2*time.Second, func() {
						completed = true
						onComplete(false)
						a.Quit()
					})
					return
				}

				passwordEntry.SetText("")
				if remainingAttempts > 1 {
					passwordEntry.SetPlaceHolder(fmt.Sprintf("Incorrect Password. %d attempts remaining", remainingAttempts))
				} else {
					passwordEntry.SetPlaceHolder("Incorrect Password. 1 attempt remaining")
				}
			}
		})

		w.SetContent(container.NewVBox(
			container.NewCenter(widget.NewRichTextFromMarkdown("# PassPort")),
			container.NewStack(img),
			container.NewCenter(widget.NewRichTextFromMarkdown("## Please enter your master password to continue:")),
			passwordEntry,
			submitButton,
		))
	} else {

		if FileExists(filename) && filename != "" {
			os.Remove(filename)
		}

		password := widget.NewPasswordEntry()
		confirm := widget.NewPasswordEntry()
		confirm.SetPlaceHolder("Confirm Master Password")

		saveBtn := widget.NewButton("Save", func() {
			if password.Text == "" || password.Text != confirm.Text {
				dialog.ShowError(fmt.Errorf("Passwords do not match or are empty"), w)
			} else {
				if err := setupMasterPasswordGUI(password.Text, confirm.Text); err != nil {
					dialog.ShowError(fmt.Errorf("Failed to set up master password: %v", err), w)
					completed = true
					onComplete(false)
					return
				}
				completed = true
				w.Close()
				onComplete(true)
			}
		})

		w.SetContent(container.NewVBox(
			container.NewCenter(widget.NewRichTextFromMarkdown("# PassPort")),
			container.NewStack(img),
			container.NewCenter(widget.NewRichTextFromMarkdown("## Welcome to PassPort! It looks like this is your first time using the application. Let's set up your master password to get started.")),
			container.NewCenter(widget.NewRichTextFromMarkdown("### Setup your master password:")),
			password,
			confirm,
			saveBtn,
		))

	}
	w.Show()
}

func authenticate(password string) bool {
	path, err := getMasterPasswordPath()
	if err != nil {
		fmt.Println("authentication failed")
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("authentication failed")
		return false
	}

	storedHash := string(data)

	bytePassword := []byte(password)

	match, err := VerifyPassword(bytePassword, storedHash)
	if err != nil {
		fmt.Println("authentication failed")
		return false
	}

	if !match {
		fmt.Println("authentication failed")
		return false
	}

	return true
}

func saveConfig(theme string) error {

	data, err := os.Open(getConfigPath())
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer data.Close()

	type Config struct {
		Theme string `json:"theme"`
	}
	var cfg Config

	if err := json.NewDecoder(data).Decode(&cfg); err != nil {
		return fmt.Errorf("error decoding config file: %v", err)
	}

	cfg.Theme = theme

	configFile, err := os.Create(getConfigPath())
	if err != nil {
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer configFile.Close()
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("error encoding config file: %v", err)
	}
	return nil

}

func gui() {

	start := time.Now()

	path := getConfigPath()

	type Config struct {
		Theme string `json:"theme"`
	}

	var cfg Config
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error opening config file")
			cfg.Theme = "DarkTheme" // default
		} else {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(&cfg); err != nil {
				fmt.Println("Failed to read config file, using default config")
				cfg.Theme = "DarkTheme" // default
			}
		}
	} else {
		cfg.Theme = "DarkTheme"
	}

	a := app.New()
	icon := fyne.NewStaticResource("logo.png", resourceLogoPngData)
	a.SetIcon(icon)
	selectedTheme := getThemeByName(cfg.Theme)
	a.Settings().SetTheme(selectedTheme)

	authWindow(a, func(authenticated bool) {
		if !authenticated {
			a.Quit()
			return
		}

		w := a.NewWindow("PassPort")

		w.Resize(fyne.NewSize(1060, 680))

		var refreshUI func()
		refreshUI = func() {
			passwords, err := checkPasswords()
			if err != nil && err.Error() != "no passwords saved" && len(passwords) != 0 {
				dialog.ShowError(fmt.Errorf("Error reading saved password: %s", err), w)
				passwords = []Password{}
			}

			containers := container.NewVBox()
			title := widget.NewRichTextFromMarkdown("# Saved Passwords: ")
			titleContainer := container.NewVBox(title)
			titleCenter := container.NewCenter(titleContainer)
			newButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
				addPasswordGUI(passwords, w, refreshUI, a, start)
			})
			editButton := widget.NewButton("Edit master password", func() {
				editMasterPasswordGUI(w, func(success bool) {
					if success {
						dialog.ShowInformation("Success", "Master password updated successfully", w)
					} else {
						dialog.ShowError(fmt.Errorf("Failed to update master password"), w)
					}
				}, a, start)
			})
			modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas()) //dummy modal to be replaced
			settingsContent := container.NewVBox(
				widget.NewLabel("Select Theme:"),
				widget.NewButton("Custom Theme", func() {
					a.Settings().SetTheme(&CustomTheme{})
					cfg.Theme = "CustomTheme"
					saveConfig(cfg.Theme)
					refreshUI()
				}),
				widget.NewButton("Dark Theme", func() {
					a.Settings().SetTheme(&DarkTheme{})
					cfg.Theme = "DarkTheme"
					saveConfig(cfg.Theme)
					refreshUI()
				}),
				widget.NewButton("Light Theme", func() {
					a.Settings().SetTheme(&LightTheme{})
					cfg.Theme = "LightTheme"
					saveConfig(cfg.Theme)
					refreshUI()
				}),
				widget.NewLabel("Change master password: "),
				editButton,
				widget.NewLabel("Logout"),
				widget.NewButton("Logout", func() {
					dialog.ShowConfirm("Confirm Logout", "Are you sure you want to logout?", func(confirmed bool) {
						if confirmed {
							w.Hide()
							authWindow(a, func(authenticated bool) {
								if authenticated {
									w.Show()
									refreshUI()
								} else {
									a.Quit()
								}
							})
						}
					}, w)
				}),
				widget.NewButton("Close", func() {
					modal.Hide()
				}),
				widget.NewLabel("PassPort v0.1.0 developed by Buct0r"), //TODO: Update at every release
			)
			settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
				modal = widget.NewModalPopUp(settingsContent, w.Canvas())
				modal.Show()
			})

			titleContainer.Add(newButton)
			titleContainer.Add(settingsButton)
			containers.Add(titleCenter)
			//buttonsContainer := container.NewHBox()
			//containers.Add(buttonsContainer)
			cardContainer := container.NewGridWrap(fyne.NewSize(300, 200))
			for _, p := range passwords {
				pwd := p
				card := widget.NewCard(pwd.Service, "Username: "+pwd.Username, container.NewVBox(
					widget.NewRichTextFromMarkdown("**Created: "+pwd.Created+"**"),
					container.NewHBox(widget.NewButton("Reveal", func() {
						password := pwd.Password
						passwordBytes := []byte(password)
						modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas()) //dummy modal to be replaced

						content := container.NewHBox(widget.NewLabel("Password: "+password), widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
							a.Clipboard().SetContent(password)
						}), widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
							modal.Hide()
							zeroBytes(passwordBytes)
							a.Clipboard().SetContent("")
						}))

						modal = widget.NewModalPopUp(content, w.Canvas())

						if time.Since(start) > 15*time.Minute {
							dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
							authWindow(a, func(authenticated bool) {
								if authenticated {
									modal.Show()
									time.AfterFunc(5*time.Second, func() {
										modal.Hide()
										zeroBytes(passwordBytes)
										a.Clipboard().SetContent("") //TODO: Make this a feature that can be toggled
									})
								} else {
									a.Quit()
								}
							})
						} else {
							// Session still valid, show password immediately
							modal.Show()
							time.AfterFunc(5*time.Second, func() {
								modal.Hide()
								zeroBytes(passwordBytes)
								a.Clipboard().SetContent("")
							})
						}

					}),
						widget.NewButton("Delete", func() {
							deletePasswordGUI(passwords, w, pwd.Service, refreshUI, a, start)
						}),
						widget.NewButton("Edit", func() {
							editPasswordGUI(passwords, w, pwd.Service, refreshUI, a, start)
						}),
					),
				))
				cardBg := getCardBackground(cfg.Theme)
				cardBg.CornerRadius = 12
				cardWithBg := container.New(layout.NewStackLayout(), cardBg, card)
				cardContainer.Add(cardWithBg)
			}
			containers.Add(cardContainer)
			w.SetContent(container.NewVScroll(containers))
		}

		refreshUI()
		w.Show()
	})

	a.Run()
}

func getThemeByName(name string) fyne.Theme {
	switch name {
	case "CustomTheme":
		return &CustomTheme{}
	case "DarkTheme":
		return &DarkTheme{}
	case "LightTheme":
		return &LightTheme{}
	default:
		return &CustomTheme{}
	}
}

func deletePasswordGUI(passwords []Password, w fyne.Window, service string, onDelete func(), a fyne.App, start time.Time) error {
	if time.Since(start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				deletePasswordGUI(passwords, w, service, onDelete, a, start)
			} else {
				a.Quit()
			}
		})
	}
	var passwordToDelete *Password
	var updatedPasswords []Password
	for _, p := range passwords {
		if p.Service == service {
			passwordToDelete = &p
		} else {
			updatedPasswords = append(updatedPasswords, p)
		}
	}

	// Check if service was found
	if passwordToDelete == nil {
		return fmt.Errorf("service '%s' not found", service)
	}

	path, err := getMasterPasswordPath()
	if err != nil {
		return fmt.Errorf("error determining master password path: %v", err)
	}

	confirmDialog := dialog.NewConfirm(
		"Confirm Deletion",
		fmt.Sprintf("Are you sure you want to delete the password for '%s'?", service),
		func(confirmed bool) {
			if confirmed {
				data, err := json.MarshalIndent(updatedPasswords, "", "  ")
				if err != nil {
					fmt.Println("Error marshaling data:", err)
					return
				}
				text, err := encryptData(filename, path, string(data))
				if err != nil {
					fmt.Println("Error encrypting data:", err)
					return
				}
				err = os.WriteFile(filename, []byte(text), 0600)
				if err != nil {
					fmt.Println("Error writing file:", err)
					return
				}

				onDelete()
			}
		},
		w,
	)
	confirmDialog.Show()

	return nil
}

func editPasswordGUI(passwords []Password, w fyne.Window, service string, onEdit func(), a fyne.App, start time.Time) error {
	if time.Since(start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				editPasswordGUI(passwords, w, service, onEdit, a, start)
			} else {
				a.Quit()
			}
		})
	}

	var passwordToEdit *Password
	for _, p := range passwords {
		if p.Service == service {
			passwordToEdit = &p
			break
		}
	}

	if passwordToEdit == nil {
		return fmt.Errorf("service '%s' not found", service)
	}

	modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas())

	serviceEntry := widget.NewEntry()
	serviceEntry.SetText(passwordToEdit.Service)
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(passwordToEdit.Username)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(passwordToEdit.Password)

	saveButton := widget.NewButton("Save", func() {
		updatedService := serviceEntry.Text
		updatedUsername := usernameEntry.Text
		updatedPassword := passwordEntry.Text
		updatedDate := time.Now().Format("2006-01-02 15:04:05")
		if updatedService == "" || updatedUsername == "" || updatedPassword == "" {
			dialog.ShowInformation("Error", "All fields must be filled out", w)
			return
		}
		passwordToEdit.Service = updatedService
		passwordToEdit.Username = updatedUsername
		passwordToEdit.Password = updatedPassword
		passwordToEdit.Created = updatedDate
		var updatedPasswords []Password
		for _, p := range passwords {
			if p.Service == service {
				updatedPasswords = append(updatedPasswords, *passwordToEdit)
			} else {
				updatedPasswords = append(updatedPasswords, p)
			}
		}
		data, err := json.MarshalIndent(updatedPasswords, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling data:", err)
			return
		}
		path, err := getMasterPasswordPath()
		if err != nil {
			fmt.Println("Error determining master password path:", err)
			return
		}
		text, err := encryptData(filename, path, string(data))
		if err != nil {
			fmt.Println("Error encrypting data:", err)
			return
		}
		err = os.WriteFile(filename, []byte(text), 0600)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		onEdit()
		modal.Hide()
	})

	content := container.NewVBox(
		widget.NewLabel("Edit Password Entry"),
		widget.NewLabel("Service:"),
		serviceEntry,
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		saveButton,
		widget.NewButton("Close", func() {
			modal.Hide()
		}),
	)

	modal = widget.NewModalPopUp(content, w.Canvas())
	modal.Show()

	return nil
}

func addPasswordGUI(passwords []Password, w fyne.Window, onAdd func(), a fyne.App, start time.Time) {
	if time.Since(start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				addPasswordGUI(passwords, w, onAdd, a, start)
			} else {
				a.Quit()
			}
		})
	}

	var newPassword Password
	modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas())

	serviceEntry := widget.NewEntry()
	serviceEntry.SetText(newPassword.Service)
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(newPassword.Username)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(newPassword.Password)
	date := time.Now().Format("2006-01-02 15:04:05")

	saveButton := widget.NewButton("Save", func() {
		updatedService := serviceEntry.Text
		updatedUsername := usernameEntry.Text
		updatedPassword := passwordEntry.Text
		if updatedService == "" || updatedUsername == "" || updatedPassword == "" {
			dialog.ShowInformation("Error", "All fields must be filled out", w)
			return
		}
		newPassword.Service = updatedService
		newPassword.Username = updatedUsername
		newPassword.Password = updatedPassword
		newPassword.Created = date
		var updatedPasswords []Password
		updatedPasswords = append(updatedPasswords, newPassword)
		for _, p := range passwords {
			updatedPasswords = append(updatedPasswords, p)
		}
		data, err := json.MarshalIndent(updatedPasswords, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling data:", err)
			return
		}
		path, err := getMasterPasswordPath()
		if err != nil {
			fmt.Println("Error determining master password path:", err)
			return
		}
		text, err := encryptData(filename, path, string(data))
		if err != nil {
			fmt.Println("Error encrypting data:", err)
			return
		}
		err = os.WriteFile(filename, []byte(text), 0600)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		onAdd()
		modal.Hide()
	})

	closeButton := widget.NewButton("Close", func() {
		modal.Hide()
	})

	content := container.NewVBox(
		widget.NewLabel("Add New Password Entry"),
		widget.NewLabel("Service:"),
		serviceEntry,
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		saveButton,
		closeButton,
	)

	modal = widget.NewModalPopUp(content, w.Canvas())
	modal.Show()
}

func setupMasterPasswordGUI(password string, confirmation string) error {

	passwordN := bytes.TrimSpace([]byte(password))
	confirmationN := bytes.TrimSpace([]byte(confirmation))
	for len(password) < 12 {
		return fmt.Errorf("master password must be at least 12 characters")
	}

	if !bytes.Equal(passwordN, confirmationN) {
		return fmt.Errorf("passwords do %snot%s match", "\033[31m", "\033[0m")
	}

	for i := range confirmationN {
		confirmationN[i] = 0
	}

	hashedPassword, err := HashPassword(passwordN)
	if err != nil {
		return err
	}
	//filename = rand.Text()
	path, err := getMasterPasswordPath()
	if err != nil {
		return fmt.Errorf("error determining master password path: %v", err)
	}

	return os.WriteFile(path, []byte(hashedPassword), 0600)

}

func editMasterPasswordGUI(w fyne.Window, onComplete func(bool), a fyne.App, start time.Time) error {

	if time.Since(start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				editMasterPasswordGUI(w, onComplete, a, start)
			} else {
				a.Quit()
			}
		})
	}

	path, err := getMasterPasswordPath()
	if err != nil {
		return fmt.Errorf("error determining master password path: %v", err)
	}

	if !FileExists(path) {

		modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas())

		passwordEntry := widget.NewPasswordEntry()
		confirmationEntry := widget.NewPasswordEntry()

		saveButton := widget.NewButton("Save", func() {
			newPassword := bytes.TrimSpace([]byte(passwordEntry.Text))
			confirmation := bytes.TrimSpace([]byte(confirmationEntry.Text))
			if newPassword == nil || confirmation == nil || len(newPassword) < 12 {
				dialog.ShowInformation("Error", "Master password must be at least 12 characters", w)
				return
			}
			if !bytes.Equal(newPassword, confirmation) {
				dialog.ShowInformation("Error", "Passwords do not match", w)
				return
			}

			for i := range confirmation {
				confirmation[i] = 0
			}

			hashedPassword, err := HashPassword(newPassword)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to hash password", w)
				return
			}

			//filename = rand.Text()

			os.WriteFile(path, []byte(hashedPassword), 0600)

			onComplete(true)
			modal.Hide()
		})

		content := container.NewVBox(
			widget.NewLabel("Edit Master Password"),
			widget.NewLabel("New Password:"),
			passwordEntry,
			widget.NewLabel("Confirm Password:"),
			confirmationEntry,
			saveButton,
		)

		modal = widget.NewModalPopUp(content, w.Canvas())
		modal.Show()

		return nil
	} else {
		path, err := getMasterPasswordPath()
		if err != nil {
			return fmt.Errorf("error determining master password path: %v", err)
		}
		data, error := decryptData(path, filename)
		if error != nil {
			return fmt.Errorf("failed to decrypt data: %v", error)
		}
		modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas())

		passwordEntry := widget.NewPasswordEntry()
		confirmationEntry := widget.NewPasswordEntry()

		text := []byte(data)
		var passwords []Password
		err = json.Unmarshal(text, &passwords)
		if err != nil {
			return fmt.Errorf("failed to unmarshal data: %v", err)
		}

		saveButton := widget.NewButton("Save", func() {

			newPassword := bytes.TrimSpace([]byte(passwordEntry.Text))
			confirmation := bytes.TrimSpace([]byte(confirmationEntry.Text))
			if newPassword == nil || confirmation == nil || len(newPassword) < 12 {
				dialog.ShowInformation("Error", "Master password must be at least 12 characters", w)
				return
			}
			if !bytes.Equal(newPassword, confirmation) {
				dialog.ShowInformation("Error", "Passwords do not match", w)
				return
			}

			for i := range confirmation {
				confirmation[i] = 0
			}

			hashedPassword, err := HashPassword(newPassword)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to hash password", w)
				return
			}

			//filename = rand.Text()

			os.WriteFile(path, []byte(hashedPassword), 0600)

			jsonData, err := json.MarshalIndent(passwords, "", "  ")
			if err != nil {
				dialog.ShowInformation("Error", "Failed to marshal data: "+err.Error(), w)
				return
			}

			encText, err := encryptData(filename, path, string(jsonData))
			if err != nil {
				dialog.ShowInformation("Error", "Failed to encrypt data: "+err.Error(), w)
				return
			}

			err = os.WriteFile(filename, []byte(encText), 0600)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to write data: "+err.Error(), w)
				return
			}

			onComplete(true)
			modal.Hide()
		})

		content := container.NewVBox(
			widget.NewLabel("Edit Master Password"),
			widget.NewLabel("New Password:"),
			passwordEntry,
			widget.NewLabel("Confirm Password:"),
			confirmationEntry,
			saveButton,
			widget.NewButton("Close", func() {
				modal.Hide()
				onComplete(false)
			}),
		)

		modal = widget.NewModalPopUp(content, w.Canvas())
		modal.Show()

		return nil

	}
}
