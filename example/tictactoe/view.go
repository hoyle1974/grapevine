package main

import (
	"fmt"
	"time"

	"github.com/hoyle1974/grapevine/client"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var app *tview.Application
var grid *tview.Grid
var gameData *tview.TreeView
var gameDataRoot *tview.TreeNode
var mongers *tview.TreeNode
var logs *tview.TextView

type gameplay struct {
	view *tview.TextView
}

func (gp *gameplay) SetPrompt(msg string) {
	gp.view.SetText(msg)
}

func setTuiApp() Gameplay {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	menu := newPrimitive("Menu")
	// main := newPrimitive("Main content")
	//sideBar := newPrimitive("Side Bar")
	gameDataRoot = tview.NewTreeNode("root")
	gameData = tview.NewTreeView().SetRoot(gameDataRoot)

	mongers = tview.NewTreeNode("mongers")
	gameDataRoot.AddChild(mongers)

	main := tview.NewTextView().
		SetDynamicColors(true)

	gp := &gameplay{main}

	logs := tview.NewTextView().
		SetDynamicColors(true)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	adapter := LogAdapter{out: logs}
	writer := tview.ANSIWriter(adapter)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: writer})

	grid = tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
		AddItem(logs, 2, 0, 2, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false).
		AddItem(gameData, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
		AddItem(main, 1, 1, 1, 1, 0, 100, false).
		AddItem(gameData, 1, 2, 1, 1, 0, 100, false)

	app = tview.NewApplication()

	return gp
}

func startTuiApp(grapevine client.Grapevine) {

	go func() {
		for {
			time.Sleep(time.Second)

			mongers.ClearChildren()
			for _, addr := range grapevine.GetMongers() {
				mongers.AddChild(tview.NewTreeNode(fmt.Sprintf("%v", addr)))
			}

			app.Draw()
		}
	}()

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
