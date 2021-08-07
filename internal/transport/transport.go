// Package transport handles sending and receiving using wormhole-william
package transport

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/psanford/wormhole-william/wormhole"
)

// Client defines the client for handling sending and receiving using wormhole-william
type Client struct {
	wormhole.Client

	// Save a reference to the window to avoid creating a new one when sending and receiving text
	display *textDisplay

	// Notification holds the settings value for if we have notifications enabled or not.
	Notifications bool

	// Verify holds information about if verification should be enabled or not.
	Verify bool

	// OverwriteExisting holds the settings value for if we should overwrite already existing files.
	OverwriteExisting bool

	// DownloadPath holds the download path used for saving received files.
	DownloadPath string
}

// ShowNotification sends a notification if c.Notifications is true.
func (c *Client) ShowNotification(title, content string) {
	if c.Notifications {
		fyne.CurrentApp().SendNotification(&fyne.Notification{Title: title, Content: content})
	}
}

// VerifySend asks the user to verify that the verification code is correct.
func (c *Client) VerifySend(input string) bool {
	verified := make(chan bool)
	dialog.ShowConfirm("Verify validation", "Is "+input+" the validation code?", func(accept bool) {
		verified <- accept
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	return <-verified
}

// VerifyRecv shows the verification code that should be compared on the other end.
func (c *Client) VerifyRecv(input string) bool {
	dialog.ShowInformation("Validation", "Your validation code is: "+input, fyne.CurrentApp().Driver().AllWindows()[0])
	return true
}

// NewClient returns a new client for sending and receiving using wormhole-william
func NewClient() *Client {
	return &Client{display: createTextWindow()}
}

// UserDownloadsFolder returns the downloads folder corresponding to the current user.
func UserDownloadsFolder() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fyne.LogError("Could not get home dir", err)
	}

	return filepath.Join(dir, "Downloads")
}
