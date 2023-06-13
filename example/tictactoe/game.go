package main

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/client"
)

var cb *Callback

func startGame(gp Gameplay) client.Grapevine {
	ctx := client.NewCallCtxWithApp("tictactoe")

	cb = &Callback{searching: true, ctx: ctx.NewCtx("Callback"), gp: gp}
	cb.grapevine = client.NewGrapevine(cb, ctx)
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
