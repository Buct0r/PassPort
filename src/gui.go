package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedtheme "github.com/Buct0r/PassPort/internal/theme"
)

const version = "1.0.0" //TODO: keep in sync with cli/cli.go

var guiLimiter = NewRateLimiter(5, 60*time.Second) // Allow 5 attempts per minute for GUI

func authWindow(a fyne.App, onComplete func(bool)) {
	w := a.NewWindow("PassPort: Authentication")
	w.Resize(fyne.NewSize(460, 640))

	completed := false

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
	img.SetMinSize(fyne.NewSize(180, 180))

	if FileExists(path) {
		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Enter Master Password")

		submitButton := widget.NewButton("Unlock", func() {
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
						fyne.Do(func() {
							completed = true
							onComplete(false)
							a.Quit()
						})
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
		submitButton.Importance = widget.HighImportance

		passwordEntry.OnSubmitted = func(s string) {
			submitButton.OnTapped()
		}

		title := canvas.NewText("PassPort", theme.Color(theme.ColorNamePrimary))
		title.TextSize = 32
		title.TextStyle = fyne.TextStyle{Bold: true}

		subtitle := widget.NewLabel("Enter your master password to continue")
		subtitle.Alignment = fyne.TextAlignCenter

		form := container.NewVBox(
			container.NewCenter(img),
			container.NewCenter(title),
			container.NewCenter(subtitle),
			widget.NewSeparator(),
			passwordEntry,
			submitButton,
		)

		w.SetContent(container.NewPadded(container.NewCenter(form)))
	} else {

		if FileExists(filename) && filename != "" {
			os.Remove(filename)
		}

		password := widget.NewPasswordEntry()
		password.SetPlaceHolder("Master Password")
		confirm := widget.NewPasswordEntry()
		confirm.SetPlaceHolder("Confirm Master Password")

		saveBtn := widget.NewButton("Create Vault", func() {
			if password.Text == "" || password.Text != confirm.Text {
				dialog.ShowError(fmt.Errorf("Passwords do not match or are empty"), w)
			} else if len(password.Text) < 12 {
				dialog.ShowError(fmt.Errorf("Master password must be at least 12 characters"), w)
			} else {
				if err := setupMasterPasswordGUI(password.Text, confirm.Text); err != nil {
					dialog.ShowError(fmt.Errorf("Failed to set up master password: %v", err), w)
					return
				}
				completed = true
				w.Close()
				onComplete(true)
			}
		})
		saveBtn.Importance = widget.HighImportance

		title := canvas.NewText("Welcome to PassPort", theme.Color(theme.ColorNamePrimary))
		title.TextSize = 28
		title.TextStyle = fyne.TextStyle{Bold: true}

		subtitle := widget.NewLabel("This looks like your first launch. Create a master password to get started.")
		subtitle.Alignment = fyne.TextAlignCenter
		subtitle.Wrapping = fyne.TextWrapWord

		form := container.NewVBox(
			container.NewCenter(img),
			container.NewCenter(title),
			container.NewCenter(subtitle),
			widget.NewSeparator(),
			widget.NewLabel("Master Password"),
			password,
			widget.NewLabel("Confirm Master Password"),
			confirm,
			saveBtn,
		)

		w.SetContent(container.NewPadded(container.NewCenter(form)))

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

	bytePassword = nil

	return true
}

func filterPasswords(passwords []Password, term string) []Password {
	if term == "" {
		return passwords
	}
	lower := bytes.ToLower([]byte(term))
	var out []Password
	for _, p := range passwords {
		if bytes.Contains(bytes.ToLower([]byte(p.Service)), lower) {
			out = append(out, p)
		}
	}
	return out
}

func runFyneGUI() {

	start := time.Now()

	cfg := sharedtheme.LoadConfig()

	a := app.New()
	icon := fyne.NewStaticResource("logo.png", resourceLogoPngData)
	a.SetIcon(icon)
	a.Settings().SetTheme(NewUserTheme(cfg.Active()))

	authWindow(a, func(authenticated bool) {
		if !authenticated {
			a.Quit()
			return
		}

		w := a.NewWindow("PassPort")

		w.Resize(fyne.NewSize(1060, 720))

		var searchTerm string
		var refreshUI func()
		refreshUI = func() {
			th := NewUserTheme(cfg.Active())

			passwords, err := checkPasswords()
			if err != nil && err.Error() != "no passwords saved" && len(passwords) != 0 {
				dialog.ShowError(fmt.Errorf("Error reading saved password: %s", err), w)
				passwords = []Password{}
			}

			// Header: application title and settings.
			title := canvas.NewText("PassPort", th.Primary())
			title.TextSize = 26
			title.TextStyle = fyne.TextStyle{Bold: true}

			settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
				showSettings(a, w, cfg, &start, refreshUI)
			})
			header := container.NewBorder(nil, nil, title, settingsButton)

			// Toolbar: add action and live search.
			addButton := widget.NewButtonWithIcon("Add Password", theme.ContentAddIcon(), func() {
				addPasswordGUI(passwords, w, refreshUI, a, &start)
			})
			addButton.Importance = widget.HighImportance

			searchBar := widget.NewEntry()
			searchBar.SetPlaceHolder("Search by service name...")
			searchBar.Text = searchTerm

			results := container.NewStack()

			var updateResults func()
			updateResults = func() {
				th := NewUserTheme(cfg.Active())
				if len(passwords) == 0 {
					results.Objects = []fyne.CanvasObject{emptyState(th, func() {
						addPasswordGUI(passwords, w, refreshUI, a, &start)
					})}
				} else {
					filtered := filterPasswords(passwords, searchTerm)
					if len(filtered) == 0 {
						msg := canvas.NewText(fmt.Sprintf("No passwords match %q", searchTerm), th.Muted())
						msg.TextSize = 16
						results.Objects = []fyne.CanvasObject{container.NewCenter(msg)}
					} else {
						grid := container.New(newCenteredGridLayout(fyne.NewSize(300, 210)))
						for _, p := range filtered {
							pwd := p
							grid.Add(passwordCard(pwd, th, w, a, passwords, refreshUI, &start))
						}
						results.Objects = []fyne.CanvasObject{container.NewVScroll(container.NewPadded(grid))}
					}
				}
				results.Refresh()
			}

			searchBar.OnChanged = func(s string) {
				searchTerm = s
				updateResults()
			}

			toolbar := container.NewBorder(nil, nil, addButton, nil, searchBar)

			top := container.NewVBox(
				container.NewPadded(header),
				container.NewPadded(toolbar),
				widget.NewSeparator(),
			)

			updateResults()
			w.SetContent(container.NewBorder(top, nil, nil, nil, results))
		}

		refreshUI()
		w.Show()
	})

	a.Run()
}

// centeredGridLayout arranges fixed-size cells in rows that wrap to the
// container width, centering each row so the leftover space is split evenly
// between the left and right margins.
type centeredGridLayout struct {
	cell     fyne.Size
	colCount int
	rowCount int
}

func newCenteredGridLayout(cell fyne.Size) fyne.Layout {
	return &centeredGridLayout{cell: cell, colCount: 1, rowCount: 1}
}

func (g *centeredGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	padding := theme.Padding()
	g.colCount = 1
	g.rowCount = 0

	if size.Width > g.cell.Width {
		g.colCount = int(math.Floor(float64(size.Width+padding) / float64(g.cell.Width+padding)))
	}

	rowWidth := float32(g.colCount)*g.cell.Width + float32(g.colCount-1)*padding
	offset := (size.Width - rowWidth) / 2
	if offset < 0 {
		offset = 0
	}

	i, x, y := 0, offset, float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if i%g.colCount == 0 {
			g.rowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.cell)

		if (i+1)%g.colCount == 0 {
			x = offset
			y += g.cell.Height + padding
		} else {
			x += g.cell.Width + padding
		}
		i++
	}
}

func (g *centeredGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return fyne.NewSize(g.cell.Width,
		(g.cell.Height*float32(rows))+(float32(rows-1)*theme.Padding()))
}

func passwordCard(pwd Password, th *UserTheme, w fyne.Window, a fyne.App, passwords []Password, refreshUI func(), start *time.Time) fyne.CanvasObject {
	bg := canvas.NewRectangle(th.CardBackground())
	bg.CornerRadius = 12
	bg.StrokeColor = th.Separator()
	bg.StrokeWidth = 1

	service := widget.NewLabelWithStyle(pwd.Service, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	service.SizeName = theme.SizeNameSubHeadingText
	service.Truncation = fyne.TextTruncateEllipsis

	username := widget.NewLabel(pwd.Username)
	username.SizeName = theme.SizeNameCaptionText
	username.Truncation = fyne.TextTruncateEllipsis

	created := widget.NewLabel("Last edited: " + pwd.Created)
	created.SizeName = theme.SizeNameCaptionText
	created.Importance = widget.LowImportance

	revealButton := widget.NewButton("Reveal", func() {
		revealPasswordGUI(pwd, w, a, start)
	})

	editButton := widget.NewButton("Edit", func() {
		editPasswordGUI(passwords, w, pwd.Service, refreshUI, a, start)
	})

	deleteButton := widget.NewButton("Delete", func() {
		deletePasswordGUI(passwords, w, pwd.Service, refreshUI, a, start)
	})
	deleteButton.Importance = widget.DangerImportance

	body := container.NewVBox(service, username, created)
	buttons := container.NewGridWithColumns(3, revealButton, editButton, deleteButton)

	content := container.NewVBox(
		container.NewPadded(body),
		layout.NewSpacer(),
		container.NewPadded(buttons),
	)

	return container.NewStack(bg, content)
}

func emptyState(th *UserTheme, onAdd func()) fyne.CanvasObject {
	icon := widget.NewIcon(theme.ContentAddIcon())

	heading := widget.NewLabel("Your vault is empty")
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.Alignment = fyne.TextAlignCenter
	heading.SizeName = theme.SizeNameHeadingText

	sub := widget.NewLabel("Add your first password to get started.")
	sub.Alignment = fyne.TextAlignCenter
	sub.Importance = widget.LowImportance

	button := widget.NewButtonWithIcon("Add your first password", theme.ContentAddIcon(), onAdd)
	button.Importance = widget.HighImportance

	box := container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(heading),
		container.NewCenter(sub),
		container.NewCenter(button),
	)
	return container.NewCenter(box)
}

func showSettings(a fyne.App, w fyne.Window, cfg *sharedtheme.Config, start *time.Time, refreshUI func()) {
	// Appearance
	themeSelect := widget.NewSelect(cfg.ThemeNames(), func(name string) {
		if name == "" || name == cfg.Theme {
			return
		}
		cfg.Theme = name
		if err := sharedtheme.SaveConfig(cfg); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save theme: %v", err), w)
		}
		a.Settings().SetTheme(NewUserTheme(cfg.Active()))
		refreshUI()
	})
	themeSelect.SetSelected(cfg.Theme)
	themeSelect.PlaceHolder = "Select a theme"

	appearance := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle("Appearance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Choose a colour scheme. Add your own in config.json."),
		themeSelect,
	))

	// Security
	editMasterButton := widget.NewButton("Change master password", func() {
		editMasterPasswordGUI(w, func(success bool) {
			if success {
				dialog.ShowInformation("Success", "Master password updated successfully", w)
			} else {
				dialog.ShowError(fmt.Errorf("Failed to update master password"), w)
			}
		}, a, start)
	})

	security := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle("Security", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		editMasterButton,
	))

	var modal *widget.PopUp

	// Session
	logoutButton := widget.NewButton("Logout", func() {
		dialog.ShowConfirm("Confirm Logout", "Are you sure you want to logout?", func(confirmed bool) {
			if confirmed {
				modal.Hide()
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
	})

	session := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle("Session", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		logoutButton,
	))

	// About
	about := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle("About", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("PassPort v"+version+" developed by Buct0r"),
	))

	closeButton := widget.NewButton("Close", func() {
		modal.Hide()
	})

	content := container.NewVBox(appearance, security, session, about, closeButton)

	modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	modal.Resize(fyne.NewSize(460, 620))
	modal.Show()
}

func revealPasswordGUI(pwd Password, w fyne.Window, a fyne.App, start *time.Time) {
	if time.Since(*start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				*start = time.Now()
				showPasswordModal(pwd, w, a)
			} else {
				a.Quit()
			}
		})
		return
	}

	*start = time.Now()
	showPasswordModal(pwd, w, a)
}

func showPasswordModal(pwd Password, w fyne.Window, a fyne.App) {
	passwordBytes := []byte(pwd.Password)
	var modal *widget.PopUp

	passwordLabel := widget.NewLabel(pwd.Password)
	passwordLabel.TextStyle = fyne.TextStyle{Monospace: true}
	passwordLabel.Wrapping = fyne.TextWrapBreak

	copyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		a.Clipboard().SetContent(pwd.Password)
	})

	closeButton := widget.NewButton("Close", func() {
		modal.Hide()
		zeroBytes(passwordBytes)
		a.Clipboard().SetContent("")
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Password for "+pwd.Service, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		passwordLabel,
		container.NewGridWithColumns(2, copyButton, closeButton),
	)

	modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	modal.Show()

	time.AfterFunc(5*time.Second, func() {
		fyne.Do(func() {
			modal.Hide()
			zeroBytes(passwordBytes)
			a.Clipboard().SetContent("")
		})
	})
}

func deletePasswordGUI(passwords []Password, w fyne.Window, service string, onDelete func(), a fyne.App, start *time.Time) error {

	if time.Since(*start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				deletePasswordGUI(passwords, w, service, onDelete, a, start)
			} else {
				a.Quit()
			}
		})
		return nil
	}

	*start = time.Now()

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

			if time.Since(*start) > 15*time.Minute {
				dialog.ShowError(fmt.Errorf("session expired"), w)
				w.Close()
				return
			}

			if confirmed {
				data, err := json.MarshalIndent(updatedPasswords, "", "  ")
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				text, err := encryptData(filename, path, string(data))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				err = os.WriteFile(filename, []byte(text), 0600)
				if err != nil {
					dialog.ShowError(err, w)
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

func editPasswordGUI(passwords []Password, w fyne.Window, service string, onEdit func(), a fyne.App, start *time.Time) error {
	if time.Since(*start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				editPasswordGUI(passwords, w, service, onEdit, a, start)
			} else {
				a.Quit()
			}
		})
		return nil
	}

	*start = time.Now()
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

		if time.Since(*start) > 15*time.Minute {
			dialog.ShowError(fmt.Errorf("session expired"), w)
			w.Close()
			return
		}

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
			dialog.ShowError(err, w)
			return
		}
		path, err := getMasterPasswordPath()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		text, err := encryptData(filename, path, string(data))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = os.WriteFile(filename, []byte(text), 0600)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		onEdit()
		modal.Hide()
	})
	saveButton.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Edit Password Entry", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Service"),
		serviceEntry,
		widget.NewLabel("Username"),
		usernameEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		container.NewGridWithColumns(2, saveButton, widget.NewButton("Close", func() {
			modal.Hide()
		})),
	)

	modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	modal.Resize(fyne.NewSize(420, 420))
	modal.Show()

	return nil
}

func addPasswordGUI(passwords []Password, w fyne.Window, onAdd func(), a fyne.App, start *time.Time) {
	if time.Since(*start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				addPasswordGUI(passwords, w, onAdd, a, start)
			} else {
				a.Quit()
			}
		})
		return
	}

	*start = time.Now()
	var newPassword Password
	modal := widget.NewModalPopUp(widget.NewLabel(""), w.Canvas())

	serviceEntry := widget.NewEntry()
	serviceEntry.SetText(newPassword.Service)
	serviceEntry.SetPlaceHolder("e.g. GitHub")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(newPassword.Username)
	usernameEntry.SetPlaceHolder("e.g. user@example.com")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(newPassword.Password)
	date := time.Now().Format("2006-01-02 15:04:05")

	saveButton := widget.NewButton("Save", func() {

		if time.Since(*start) > 15*time.Minute {
			dialog.ShowError(fmt.Errorf("session expired"), w)
			w.Close()
			return
		}

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
			dialog.ShowError(err, w)
			return
		}
		path, err := getMasterPasswordPath()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		text, err := encryptData(filename, path, string(data))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = os.WriteFile(filename, []byte(text), 0600)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		onAdd()
		modal.Hide()
	})
	saveButton.Importance = widget.HighImportance

	closeButton := widget.NewButton("Close", func() {
		modal.Hide()
	})

	passwordEntry.OnSubmitted = func(s string) {
		saveButton.OnTapped()
	}

	content := container.NewVBox(
		widget.NewLabelWithStyle("Add New Password Entry", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Service"),
		serviceEntry,
		widget.NewLabel("Username"),
		usernameEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		container.NewGridWithColumns(2, saveButton, closeButton),
	)

	modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	modal.Resize(fyne.NewSize(420, 420))
	modal.Show()
}

func setupMasterPasswordGUI(password string, confirmation string) error {

	passwordN := bytes.TrimSpace([]byte(password))
	confirmationN := bytes.TrimSpace([]byte(confirmation))
	if len(passwordN) < 12 {
		return fmt.Errorf("master password must be at least 12 characters")
	}

	if !bytes.Equal(passwordN, confirmationN) {
		return fmt.Errorf("passwords do not match")
	}

	for i := range confirmationN {
		confirmationN[i] = 0
	}

	hashedPassword, err := HashPassword(passwordN)
	if err != nil {
		return fmt.Errorf("Failed to hash password")
	}

	path, err := getMasterPasswordPath()
	if err != nil {
		return fmt.Errorf("Failed to determine master password path: %v", err)
	}

	return os.WriteFile(path, []byte(hashedPassword), 0600)

}

func editMasterPasswordGUI(w fyne.Window, onComplete func(bool), a fyne.App, start *time.Time) error {

	if time.Since(*start) > 15*time.Minute {
		dialog.ShowInformation("Session Expired", "Your session has expired due to inactivity. Please re-authenticate to continue.", w)
		authWindow(a, func(authenticated bool) {
			if authenticated {
				editMasterPasswordGUI(w, onComplete, a, start)
			} else {
				a.Quit()
			}
		})
		return nil
	}

	*start = time.Now()
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
			if len(newPassword) == 0 || len(confirmation) == 0 || len(newPassword) < 12 {
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

			err = os.WriteFile(path, []byte(hashedPassword), 0600)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to write master password", w)
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
		)

		modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
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
			if len(newPassword) == 0 || len(confirmation) == 0 || len(newPassword) < 12 {
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

			oldKeyData, err := os.ReadFile(path)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to backup master password: %v", err), w)
				return
			}
			oldSecretData, err := os.ReadFile(filename)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to backup secret file: %v", err), w)
				return
			}

			dataDir := filepath.Dir(path)
			timestamp := time.Now().Format("02-01-2006-150405") // DD-MM-YYYY-HHMMSS
			backupKeyPath := filepath.Join(dataDir, "master.key.backup."+timestamp)
			backupSecretPath := filepath.Join(dataDir, "secret.backup."+timestamp)

			os.WriteFile(backupKeyPath, oldKeyData, 0600)
			os.WriteFile(backupSecretPath, oldSecretData, 0600)

			restoreFromBackup := func() {
				os.WriteFile(path, oldKeyData, 0600)
				os.WriteFile(filename, oldSecretData, 0600)
				os.Remove(backupKeyPath)
				os.Remove(backupSecretPath)
			}

			err = os.WriteFile(path, []byte(hashedPassword), 0600)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to write master password", w)
				os.Remove(backupKeyPath)
				os.Remove(backupSecretPath)
				return
			}

			jsonData, err := json.MarshalIndent(passwords, "", "  ")
			if err != nil {
				restoreFromBackup()
				dialog.ShowInformation("Error", "Failed to marshal data: "+err.Error(), w)
				return
			}

			encText, err := encryptData(filename, path, string(jsonData))
			if err != nil {
				restoreFromBackup()
				dialog.ShowInformation("Error", "Failed to encrypt data: "+err.Error(), w)
				return
			}

			tmpSecret := filename + ".tmp"
			err = os.WriteFile(tmpSecret, []byte(encText), 0600)
			if err != nil {
				restoreFromBackup()
				os.Remove(tmpSecret)
				dialog.ShowInformation("Error", "Failed to write data: "+err.Error(), w)
				return
			}

			err = os.Rename(tmpSecret, filename)
			if err != nil {
				restoreFromBackup()
				os.Remove(tmpSecret)
				dialog.ShowInformation("Error", "Failed to write data: "+err.Error(), w)
				return
			}

			os.Remove(backupKeyPath)
			os.Remove(backupSecretPath)

			onComplete(true)
			modal.Hide()
		})

		content := container.NewVBox(
			widget.NewLabelWithStyle("Edit Master Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("New Password"),
			passwordEntry,
			widget.NewLabel("Confirm Password"),
			confirmationEntry,
			container.NewGridWithColumns(2, saveButton, widget.NewButton("Close", func() {
				modal.Hide()
			})),
		)

		modal = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
		modal.Resize(fyne.NewSize(420, 360))
		modal.Show()

		return nil

	}
}
