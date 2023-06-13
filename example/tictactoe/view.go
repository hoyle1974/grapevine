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

var b00, b10, b20, b01, b11, b21, b02, b12, b22 *tview.Button

type gameplay struct {
	view *tview.TextView
}

func (gp *gameplay) SetPrompt(msg string) {
	gp.view.SetText(msg)
}

func (gp *gameplay) UpdateBoard(b string) {
	if len(b) != 9 || b22 == nil {
		return
	}

	b00.SetLabel(fmt.Sprintf("%c", b[0]))
	b10.SetLabel(fmt.Sprintf("%c", b[1]))
	b20.SetLabel(fmt.Sprintf("%c", b[2]))
	b01.SetLabel(fmt.Sprintf("%c", b[3]))
	b11.SetLabel(fmt.Sprintf("%c", b[4]))
	b21.SetLabel(fmt.Sprintf("%c", b[5]))
	b02.SetLabel(fmt.Sprintf("%c", b[6]))
	b12.SetLabel(fmt.Sprintf("%c", b[7]))
	b22.SetLabel(fmt.Sprintf("%c", b[8]))
}

func setTuiApp(gi GameInput) Gameplay {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	gameDataRoot = tview.NewTreeNode("root")
	gameData = tview.NewTreeView().SetRoot(gameDataRoot)

	mongers = tview.NewTreeNode("mongers")
	gameDataRoot.AddChild(mongers)

	main := tview.NewTextView().
		SetDynamicColors(true)

	gp := &gameplay{main}

	b00 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(0, 0)
	})
	b10 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(1, 0)
	})
	b20 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(2, 0)
	})
	b01 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(0, 1)
	})
	b11 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(1, 1)
	})
	b21 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(2, 1)
	})
	b02 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(0, 2)
	})
	b12 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(1, 2)
	})
	b22 = tview.NewButton(".").SetSelectedFunc(func() {
		gi.Click(2, 2)
	})

	menu := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(b00, 0, 1, false).
			AddItem(b10, 0, 1, false).
			AddItem(b20, 0, 1, false), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(b01, 0, 1, false).
			AddItem(b11, 0, 1, false).
			AddItem(b21, 0, 1, false), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(b02, 0, 1, false).
			AddItem(b12, 0, 1, false).
			AddItem(b22, 0, 1, false), 0, 1, false)

	logs := tview.NewTextView().
		SetDynamicColors(true)

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
			time.Sleep(time.Second / 30)

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
