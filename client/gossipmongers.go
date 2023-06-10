package client

import (
	"math/rand"
	"sync"
	"time"

	"github.com/hoyle1974/grapevine/common"
)

type GossipMongers interface {
	AddMonger(addr common.Address)
	RemoveMonger(addr common.Address)
	GetRandomServerAddress() *common.Address
}

type monger struct {
	expiry        time.Time
	serverAddress common.Address
}

type gossipMongers struct {
	lock    sync.RWMutex
	ctx     CallCtx
	mongers []monger
	self    common.Address
}

func NewGossipMongers(ctx CallCtx, self common.Address) GossipMongers {
	return &gossipMongers{ctx: ctx, mongers: []monger{}, self: self}
}

func (g *gossipMongers) RemoveMonger(addr common.Address) {
	// log := g.ctx.NewCtx("RemoveMonger")
	if g.self.Equal(addr) {
		// log.Info().Msg("Not adding ourself")
		return
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	for idx, m := range g.mongers {
		// We know about this server, upgrade it's expiry
		if m.serverAddress.Equal(addr) {
			g.mongers = append(g.mongers[0:idx], g.mongers[idx+1:len(g.mongers)]...)
		}
	}

}

func (g *gossipMongers) AddMonger(addr common.Address) {
	log := g.ctx.NewCtx("AddMonger")
	if g.self.Equal(addr) {
		// log.Info().Msg("Not adding ourself")
		return
	}

	g.lock.RLock()

	for idx, m := range g.mongers {
		// We know about this server, upgrade it's expiry
		if m.serverAddress.Equal(addr) {
			// log.Info().Msg("Already know about server, upgrading expiry")
			g.lock.RUnlock()
			g.lock.Lock()
			m.expiry = time.Now().Add(time.Hour)
			g.mongers[idx] = m
			g.lock.Unlock()
			return
		}
	}
	g.lock.RUnlock()
	g.lock.Lock()
	log.Info().Msgf("Add new monger,  %v", addr.String())

	g.mongers = append(g.mongers, monger{serverAddress: addr, expiry: time.Now().Add(time.Hour)})
	g.lock.Unlock()
}

func (g *gossipMongers) GetRandomServerAddress() *common.Address {
	log := g.ctx.NewCtx("GetRandomServerAddress")

	g.lock.RLock()
	defer g.lock.RUnlock()

	log.Debug().Msgf("Monger List Size: %d", len(g.mongers))

	if len(g.mongers) == 0 {
		return nil
	}

	return &g.mongers[rand.Intn(len(g.mongers))].serverAddress
}
