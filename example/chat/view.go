package main

import (
	"github.com/rs/zerolog/log"

	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

var app *tview.Application
var grid *tview.Grid
var gameData *tview.TreeView
var gameDataRoot *tview.TreeNode
var mongers *tview.TreeNode
var logs *tview.TextView

func newTuiApp() *tview.Application {

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	main := tview.NewTextView().SetDynamicColors(true)

	logs := tview.NewTextView().SetDynamicColors(true)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	adapter := LogAdapter{out: logs}
	//writer := tview.ANSIWriter(adapter)
	//log.Logger = log.Output(zerolog.ConsoleWriter{Out: writer})
	log.Logger = log.Output(adapter)

	grid = tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
		AddItem(logs, 2, 0, 2, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	//grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
	grid.AddItem(main, 1, 0, 1, 3, 0, 0, false)

	// // Layout for screens wider than 100 cells.
	// grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
	// 	AddItem(main, 1, 1, 1, 1, 0, 100, false).
	// 	AddItem(gameData, 1, 2, 1, 1, 0, 100, false)

	return tview.NewApplication()
}
