package main

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/common"
	"github.com/hoyle1974/grapevine/grapevine"
)

type GameInput interface {
	Click(x int, y int)
	Chat(msg string)
}

var cb *Callback

func initGame() *Callback {
	ctx := common.NewCallCtxWithApp("tictactoe")

	cb = &Callback{searching: true, ctx: ctx.NewCtx("Callback")}

	return cb
}

func startGame(cb *Callback) grapevine.Grapevine {
	ctx := common.NewCallCtxWithApp("tictactoe")

	cb.grapevine = grapevine.NewGrapevine(cb, ctx)
	ip := GetOutboundIP(ctx)
	ctx.Info().Msgf("Outbound IP is: %v", ip)
	port, err := cb.grapevine.Start(ip)
	if err != nil {
		ctx.Error().Err(err).Msg("Error starting grapevine")
	}

	username := fmt.Sprintf("U%d", rand.Int()%1000000)
	password := "P" + uuid.New().String()
	ctx.Info().Msgf("Creating account with User %v and Password %v", username, password)
	err = cb.grapevine.CreateAccount(username, password)
	if err != nil {
		ctx.Error().Err(err).Msg("Error creating account")
		return nil
	}

	ctx.Info().Msg("Logging in")
	accountId, err := cb.grapevine.Login(username, password, ip, port)
	if err != nil {
		ctx.Error().Err(err).Msg("Error logging in")
		return nil
	}
	ctx.Info().Msgf("Logged in to account: %s", accountId.String())

	cb.grapevine.Search(gameType)

	return cb.grapevine
}
