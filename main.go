package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/Jacalz/rymdport/v3/internal/assets"
	"github.com/Jacalz/rymdport/v3/internal/ui"
	"github.com/Jacalz/rymdport/v3/internal/util"
	"github.com/rymdport/easypgo"
)

func main() {
	stop := easypgo.Generate()
	defer stop()

	removeTmpCrashDump := util.SetUpCrashLogging()

	a := app.NewWithID("io.github.jacalz.rymdport")
	assets.SetIcon(a)
	w := a.NewWindow("Rymdport")

	w.SetContent(ui.Create(a, w))
	w.Resize(fyne.NewSize(700, 400))
	w.SetMaster()
	w.ShowAndRun()

	removeTmpCrashDump() // Can't be deferred because we don't want it to run if we panic.
}
