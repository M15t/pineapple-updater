package main

import (
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/cavaliergopher/grab/v3"
)

func mainUI(versions []int, links map[int]string) fyne.CanvasObject {
	var selVersion int

	// * define right side
	rightSide := container.New(
		layout.NewVBoxLayout(),
		widget.NewButton("Download", func() {
			if err := download(links[selVersion]); err != nil {
				errorUI(err.Error())
			}
		}),
		widget.NewButton("Remove", func() {
			// TODO: add remove function
		}),
	)

	// * define topper
	currentPathLbl := widget.NewLabel("Current path: " + fyne.CurrentApp().Preferences().StringWithFallback("path", DefaultPath))
	topper := container.New(
		layout.NewHBoxLayout(),
		currentPathLbl,
	)

	// * define footer
	footer := container.New(
		layout.NewHBoxLayout(),
		widget.NewButtonWithIcon("", resourceIconPng, func() { go aboutUI() }), // About button
		widget.NewButton("Settings", func() { settingsUI(currentPathLbl) }),    // Settings button
	)

	// * define list of versions
	listObj := widget.NewList(
		func() int { return len(versions) },
		func() fyne.CanvasObject {
			return widget.NewLabel("EA")
		},
		func(id int, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText("EA " + strconv.Itoa(versions[id]))
		},
	)

	// * set action for selection
	listObj.OnSelected = func(id int) {
		selVersion = versions[id]
	}

	return container.New(layout.NewBorderLayout(topper, footer, nil, rightSide), topper, footer, rightSide, listObj)
}

func errorUI(message string) fyne.Window {
	a := fyne.CurrentApp()
	w := a.NewWindow("Error Window")

	// Create a label with the error message
	errorLabel := widget.NewLabel(message)

	// Create a button to close the window
	closeButton := widget.NewButton("Close", func() {
		w.Close()
	})

	// Create a container for the widgets
	content := container.NewVBox(
		container.New(layout.NewCenterLayout(), errorLabel),
		container.New(layout.NewCenterLayout(), closeButton),
	)

	// Set the content and show the window
	w.SetContent(content)
	w.Show()

	return w
}

func successUI(message string) fyne.Window {
	a := fyne.CurrentApp()
	w := a.NewWindow("Successful Window")

	// Create a label with the error message
	errorLabel := widget.NewLabel(message)

	// Create a button to close the window
	closeButton := widget.NewButton("Close", func() {
		w.Close()
	})

	// Create a container for the widgets
	content := container.NewVBox(
		container.New(layout.NewCenterLayout(), errorLabel),
		container.New(layout.NewCenterLayout(), closeButton),
	)

	// Set the content and show the window
	w.SetContent(content)
	w.Show()

	return w
}

func aboutUI() {
	a := fyne.CurrentApp()
	w := a.NewWindow("About")
	w.Resize(fyne.NewSize(400, 300))

	//TODO: set proper layout instead of using newlines in labels
	logo := canvas.NewImageFromResource(resourceIconPng)
	logo.FillMode = canvas.ImageFillOriginal
	quitButton := widget.NewButton("close", func() { w.Close() })
	aboutText1 := widget.NewLabelWithStyle("Project Dëfënëstrëring", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	aboutText2 := widget.NewLabelWithStyle("\nFrom EmuWorld with love\n2021", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	aboutText3 := widget.NewLabelWithStyle("\n\n\nThis program is free software; you can redistribute it and/or modify\nit under the terms of the GNU General Public License as published by\nthe Free Software Foundation; either version 2 of the License, or\n(at your option) any later version.", fyne.TextAlignCenter, fyne.TextStyle{})
	ui := container.New(layout.NewBorderLayout(logo, quitButton, nil, nil), logo, quitButton, aboutText1, aboutText2, aboutText3)

	w.SetIcon(resourceIconPng)
	w.SetContent(ui)
	w.SetFixedSize(true)
	w.Show()
}

// TODO: make it pretty and add ETA
func downloadUI(resp *grab.Response, cancel func()) {
	a := fyne.CurrentApp()
	w := a.NewWindow("Downloading...")
	downloadProgress := widget.NewProgressBar()
	downloadSpeed := widget.NewLabel("")

	w.Resize(fyne.NewSize(400, 200))
	w.SetIcon(resourceIconPng)
	w.SetFixedSize(true)
	w.SetOnClosed(func() { cancel() })
	w.SetContent(container.New(layout.NewBorderLayout(downloadProgress, nil, nil, nil), downloadProgress, downloadSpeed))
	w.Show()

	// Async loop to update the UI
	go func() {
		for {
			time.Sleep(time.Millisecond * 250)
			downloadProgress.SetValue(resp.Progress())
			downloadSpeed.SetText("Download Speed: " + strconv.Itoa(int(resp.BytesPerSecond()/1000)) + "KByte/s")
			if int(resp.Progress()) == 1 {
				successUI("Download completed!")
				w.Close()
				break
			}
		}
	}()
}

// TODO: make it pretty (fixed window size?) and add checkmark to create shortcuts
// TODO: update window when Path is changed
// TODO: make the file browser bigger by default (seperate window?)
func settingsUI(currentPathLbl *widget.Label) fyne.CanvasObject {
	a := fyne.CurrentApp()
	w := a.NewWindow("Settings")
	w.Resize(fyne.NewSize(500, 400))

	installPath := widget.NewLabel(a.Preferences().StringWithFallback("path", DefaultPath))
	setNewPath := widget.NewButton("Set path", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			// Update the installPath label with the new path
			newPath := setPath(uri, err)
			installPath.SetText(a.Preferences().StringWithFallback("path", newPath))
			currentPathLbl.SetText("Current path: " + newPath)
			w.Close()
		}, w)
	})

	ui := container.New(layout.NewHBoxLayout(), installPath, setNewPath)

	w.SetContent(ui)
	w.Show()

	return container.NewWithoutLayout(w.Content())
}

// remove filename from path, because fyne is dumb
func setPath(path fyne.ListableURI, err error) (newPath string) {
	n, _ := path.List()

	newPath = filepath.Dir(n[0].Path())

	fyne.CurrentApp().Preferences().SetString("path", newPath)

	return
}
