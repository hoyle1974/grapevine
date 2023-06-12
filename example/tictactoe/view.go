package main

import (
	"time"

	"github.com/rivo/tview"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var app *tview.Application
var grid *tview.Grid

func setTuiApp() {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	menu := newPrimitive("Menu")
	// main := newPrimitive("Main content")
	sideBar := newPrimitive("Side Bar")

	main := tview.NewTextView().
		SetDynamicColors(true)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	adapter := LogAdapter{out: main}
	writer := tview.ANSIWriter(adapter)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: writer})

	grid = tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
		AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false).
		AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
		AddItem(main, 1, 1, 1, 1, 0, 100, false).
		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	app = tview.NewApplication()
}

func startTuiApp() {
	go func() {
		for {
			time.Sleep(time.Second / 5)
			app.Draw()
		}
	}()

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
