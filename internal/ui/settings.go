package ui

import (
	"errors"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	appearance "fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Jacalz/rymdport/v3/internal/transport"
	"github.com/Jacalz/rymdport/v3/internal/updater"
	"github.com/Jacalz/rymdport/v3/internal/util"
	"github.com/rymdport/wormhole/wormhole"
)

type settings struct {
	downloadPathEntry *widget.Entry
	overwriteFiles    *widget.RadioGroup
	notificationRadio *widget.RadioGroup
	extractRadio      *widget.RadioGroup
	checkUpdatesRadio *widget.RadioGroup

	componentSlider     *widget.Slider
	componentLabel      *widget.Label
	verifyRadio         *widget.RadioGroup
	appID               *widget.SelectEntry
	rendezvousURL       *widget.SelectEntry
	transitRelayAddress *widget.SelectEntry

	client      *transport.Client
	preferences fyne.Preferences
	window      fyne.Window
}

func newSettingsTab(w fyne.Window, c *transport.Client) *container.TabItem {
	settings := &settings{window: w, client: c, preferences: c.App.Preferences()}

	return &container.TabItem{
		Text:    "Settings",
		Icon:    theme.SettingsIcon(),
		Content: settings.buildUI(c.App),
	}
}

func (s *settings) onDownloadsPathSubmitted(path string) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		dialog.ShowInformation("Does not exist", "Please select a valid directory.", s.window)
		return
	} else if err != nil {
		fyne.LogError("Error when trying to verify directory", err)
		dialog.ShowError(err, s.window)
		return
	} else if !info.IsDir() {
		dialog.ShowInformation("Not a directory", "Please select a valid directory.", s.window)
		return
	}

	s.client.DownloadPath = path
	s.preferences.SetString("DownloadPath", s.client.DownloadPath)
	s.downloadPathEntry.SetText(s.client.DownloadPath)
}

func (s *settings) onDownloadsPathSelected() {
	folder := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			fyne.LogError("Error on selecting folder", err)
			dialog.ShowError(err, s.window)
			return
		} else if folder == nil {
			return
		}

		s.client.DownloadPath = folder.Path()
		s.preferences.SetString("DownloadPath", s.client.DownloadPath)
		s.downloadPathEntry.SetText(s.client.DownloadPath)
	}, s.window)

	folder.Resize(util.WindowSizeToDialog(s.window.Canvas().Size()))
	folder.Show()
}

func (s *settings) onOverwriteFilesChanged(selected string) {
	if selected == "Off" {
		s.client.OverwriteExisting = false
		s.preferences.SetBool("OverwriteFiles", s.client.OverwriteExisting)
		return
	}

	confirm := dialog.NewConfirm("Are you sure?", "Enabling this option risks potentially overwriting important files.", func(enable bool) {
		if !enable {
			s.overwriteFiles.SetSelected("Off")
			return
		}

		s.client.OverwriteExisting = true
		s.preferences.SetBool("OverwriteFiles", s.client.OverwriteExisting)
	}, s.window)
	confirm.SetConfirmImportance(widget.WarningImportance)
	confirm.Show()
}

func (s *settings) onNotificationsChanged(selected string) {
	s.client.Notifications = selected == "On"
	s.preferences.SetBool("Notifications", s.client.Notifications)
}

func (s *settings) onExtractChanged(selected string) {
	s.client.NoExtractDirectory = selected == "Off" // UI representation is flipped.
	s.preferences.SetBool("NoExtractDirectory", s.client.NoExtractDirectory)
}

func (s *settings) onCheckUpdatesChanged(selected string) {
	s.preferences.SetBool("CheckUpdates", selected == "On")
}

func (s *settings) onComponentsChange(value float64) {
	s.componentLabel.SetText(string('0' + byte(value)))
}

func (s *settings) onComponentsChangeEnded(value float64) {
	s.client.PassPhraseComponentLength = int(value)
	s.preferences.SetInt("ComponentLength", int(value))
}

func (s *settings) onAppIDChanged(appID string) {
	s.client.AppID = appID
	s.preferences.SetString("AppID", appID)
}

func (s *settings) onRendezvousURLChange(url string) {
	s.client.RendezvousURL = url
	s.preferences.SetString("RendezvousURL", url)
}

func (s *settings) onTransitAdressChange(address string) {
	s.client.TransitRelayAddress = address
	s.preferences.SetString("TransitRelayAddress", address)
}

func (s *settings) onVerifyChanged(selected string) {
	enabled := selected == "On"
	s.preferences.SetBool("Verify", enabled)
	if enabled {
		s.client.VerifierOk = s.verify
	} else {
		s.client.VerifierOk = nil
	}
}

func (s *settings) verify(hash string) bool {
	verified := make(chan bool)
	dialog.ShowCustomConfirm("Verify content", "Accept", "Reject",
		container.NewVBox(
			newBoldLabel("The hash for your content is:"),
			&widget.Label{Text: hash, Wrapping: fyne.TextWrapBreak},
			newBoldLabel("Please verify that the hash is the same on both sides."),
		), func(accept bool) { verified <- accept }, s.window,
	)

	return <-verified
}

// getPreferences is used to set the preferences on startup without saving at the same time.
func (s *settings) getPreferences(app fyne.App) {
	s.client.DownloadPath = s.preferences.StringWithFallback("DownloadPath", util.UserDownloadsFolder())
	s.downloadPathEntry.Text = s.client.DownloadPath

	s.client.OverwriteExisting = s.preferences.Bool("OverwriteFiles")
	s.overwriteFiles.Selected = onOrOff(s.client.OverwriteExisting)

	s.client.Notifications = s.preferences.BoolWithFallback("Notifications", true)
	s.notificationRadio.Selected = onOrOff(s.client.Notifications)

	s.client.NoExtractDirectory = s.preferences.Bool("NoExtractDirectory")
	s.extractRadio.Selected = onOrOff(!s.client.NoExtractDirectory)

	checkUpdates := s.preferences.BoolWithFallback("CheckUpdates", true)
	if !updater.Enabled {
		checkUpdates = false
		s.checkUpdatesRadio.Disable()
	}
	if checkUpdates {
		updater.Enable(app, s.window)
	}
	s.checkUpdatesRadio.Selected = onOrOff(checkUpdates)

	verify := s.preferences.Bool("Verify")
	s.verifyRadio.Selected = onOrOff(verify)
	if verify {
		s.client.VerifierOk = s.verify
	}

	s.client.PassPhraseComponentLength = s.preferences.IntWithFallback("ComponentLength", 2)
	s.componentSlider.Value = float64(s.client.PassPhraseComponentLength)
	s.componentLabel.Text = string('0' + byte(s.componentSlider.Value))

	s.client.AppID = s.preferences.String("AppID")
	s.appID.Text = s.client.AppID

	s.client.RendezvousURL = s.preferences.String("RendezvousURL")
	s.rendezvousURL.Text = s.client.RendezvousURL

	s.client.TransitRelayAddress = s.preferences.String("TransitRelayAddress")
	s.transitRelayAddress.Text = s.client.TransitRelayAddress
}

func (s *settings) buildUI(app fyne.App) *container.Scroll {
	onOffOptions := []string{"On", "Off"}

	pathSelector := &widget.Button{Icon: theme.FolderOpenIcon(), Importance: widget.LowImportance, OnTapped: s.onDownloadsPathSelected}
	s.downloadPathEntry = &widget.Entry{Scroll: container.ScrollHorizontalOnly, OnSubmitted: s.onDownloadsPathSubmitted, ActionItem: pathSelector}

	s.overwriteFiles = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onOverwriteFilesChanged}

	s.notificationRadio = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onNotificationsChanged}

	s.extractRadio = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onExtractChanged}

	s.checkUpdatesRadio = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onCheckUpdatesChanged}

	s.verifyRadio = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onVerifyChanged}

	s.componentSlider = &widget.Slider{Min: 2.0, Max: 9.0, Step: 1, OnChanged: s.onComponentsChange, OnChangeEnded: s.onComponentsChangeEnded}
	s.componentLabel = &widget.Label{}

	s.appID = widget.NewSelectEntry([]string{wormhole.WormholeCLIAppID})
	s.appID.PlaceHolder = wormhole.WormholeCLIAppID
	s.appID.OnChanged = s.onAppIDChanged

	const leastAuthorityRendzvousURL = "wss://mailbox.mw.leastauthority.com/v1"
	s.rendezvousURL = widget.NewSelectEntry([]string{wormhole.DefaultRendezvousURL, leastAuthorityRendzvousURL})
	s.rendezvousURL.PlaceHolder = wormhole.DefaultRendezvousURL
	s.rendezvousURL.OnChanged = s.onRendezvousURLChange

	const leastAuthorityTransitRelayAddress = "relay.mw.leastauthority.com:4001"
	s.transitRelayAddress = widget.NewSelectEntry([]string{wormhole.DefaultTransitRelayAddress, leastAuthorityTransitRelayAddress})
	s.transitRelayAddress.PlaceHolder = wormhole.DefaultTransitRelayAddress
	s.transitRelayAddress.OnChanged = s.onTransitAdressChange

	s.getPreferences(app)

	interfaceContainer := appearance.NewSettings().LoadAppearanceScreen(s.window)

	dataContainer := container.NewGridWithColumns(2,
		newBoldLabel("Save files to"), s.downloadPathEntry,
		newBoldLabel("Overwrite files"), s.overwriteFiles,
		newBoldLabel("Notifications"), s.notificationRadio,
		newBoldLabel("Extract received directory"), s.extractRadio,
		newBoldLabel("Check for updates"), s.checkUpdatesRadio,
	)

	wormholeContainer := container.NewVBox(
		container.NewGridWithColumns(2,
			newBoldLabel("Verify before accepting"), s.verifyRadio,
			newBoldLabel("Passphrase length"),
			container.NewBorder(nil, nil, nil, s.componentLabel, s.componentSlider),
		),
		&widget.Accordion{Items: []*widget.AccordionItem{
			{Title: "Advanced", Detail: container.NewGridWithColumns(2,
				newBoldLabel("AppID"), s.appID,
				newBoldLabel("Rendezvous URL"), s.rendezvousURL,
				newBoldLabel("Transit Relay Address"), s.transitRelayAddress,
			)},
		}},
	)

	return container.NewScroll(container.NewVBox(
		&widget.Card{Title: "User Interface", Content: interfaceContainer},
		&widget.Card{Title: "Data Handling", Content: dataContainer},
		&widget.Card{Title: "Wormhole Options", Content: wormholeContainer},
	))
}

func newBoldLabel(text string) *widget.Label {
	return &widget.Label{Text: text, TextStyle: fyne.TextStyle{Bold: true}}
}

func onOrOff(on bool) string {
	if on {
		return "On"
	}

	return "Off"
}
