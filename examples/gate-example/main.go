package main

import (
	"context"
	"time"

	goakt "github.com/tochemey/goakt/v4/actor"
	"github.com/tochemey/goakt/v4/log"
)

type Open struct{}
type Close struct{}
type Use struct{}

type Gate struct{}

func (g *Gate) PreStart(*goakt.Context) error { return nil }

func (g *Gate) Receive(ctx *goakt.ReceiveContext) {
	// initial behavior: closed
	g.closedBehavior(ctx)
}

func (g *Gate) closedBehavior(ctx *goakt.ReceiveContext) {
	switch ctx.Message().(type) {
	case *Open:
		ctx.Become(g.openBehavior)
		ctx.Response("gate opened")
	case *Use:
		ctx.Response("cannot use: gate is closed")
	default:
		ctx.Unhandled()
	}
}

func (g *Gate) openBehavior(ctx *goakt.ReceiveContext) {
	switch ctx.Message().(type) {
	case *Close:
		ctx.UnBecome() // go back to closed behavior
		ctx.Response("gate closed")
	case *Use:
		ctx.Response("using gate...")
	default:
		ctx.Unhandled()
	}
}

func (g *Gate) PostStop(*goakt.Context) error { return nil }

func main() {
	ctx := context.Background()
	logger := log.DefaultLogger

	// Create ActorSystem
	system, err := goakt.NewActorSystem(
		"gate-system",
		goakt.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	if err := system.Start(ctx); err != nil {
		panic(err)
	}

	// Spawn the Gate actor
	gatePID, err := system.Spawn(ctx, "gate-1", &Gate{})
	if err != nil {
		panic(err)
	}

	// try to use while closed
	r1, _ := goakt.Ask(ctx, gatePID, &Use{}, time.Second)
	logger.Info("response 1", "msg", r1)

	// open it
	rOpen, _ := goakt.Ask(ctx, gatePID, &Open{}, time.Second)
	logger.Info("response open", "msg", rOpen)

	// use while open
	r2, _ := goakt.Ask(ctx, gatePID, &Use{}, time.Second)
	logger.Info("response 2", "msg", r2)

	// close again
	rClose, _ := goakt.Ask(ctx, gatePID, &Close{}, time.Second)
	logger.Info("response close", "msg", rClose)

	_ = system.Stop(ctx)
}
