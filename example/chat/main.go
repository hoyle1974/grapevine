package main

import (
	"flag"
)

func main() {
	flag.Parse()

	// Start GUI
	app := newTuiApp()

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	/*
	   gameInput := initGame()

	   gp := setTuiApp(gameInput)

	   gameInput.gp = gp

	   ctx := common.NewCallCtxWithApp("tictactoe")
	   ctx.Info().Msg("Flags:")

	   	flag.CommandLine.VisitAll(func(flag *flag.Flag) {
	   		ctx.Info().Msg(fmt.Sprintf("\t%v:%v", flag.Name, flag.Value))
	   	})

	   info, ok := debug.ReadBuildInfo()

	   	if !ok {
	   		panic("couldn't read build info")
	   	}

	   ctx.Info().Msg("Build Info Version: " + info.Main.Version + " " + info.Main.Sum)

	   grapevine := startGame(gameInput)

	   startTuiApp(grapevine)
	*/
}
