package main

import (
	"context"
	"time"

	goakt "github.com/tochemey/goakt/v4/actor"
	"github.com/tochemey/goakt/v4/log"
)

// Messages
type Greet struct{ Name string }
type Greeting struct{ Message string }

// Actor
type Greeter struct {
	greeted int
}

func (g *Greeter) PreStart(*goakt.Context) error {
	return nil
}

func (g *Greeter) Receive(ctx *goakt.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *Greet:
		g.greeted++
		ctx.Response(
			&Greeting{Message: "Hello " + msg.Name},
		)
	default:
		ctx.Unhandled()
	}
}

func (g *Greeter) PostStop(*goakt.Context) error {
	return nil
}

func main() {
	ctx := context.Background()
	logger := log.DefaultLogger

	system, _ := goakt.NewActorSystem(
		"demo",
		goakt.WithLogger(logger),
	)

	_ = system.Start(ctx)

	pid, _ := system.Spawn(
		ctx,
		"greeter",
		&Greeter{},
	)

	reply, _ := goakt.Ask(
		ctx,
		pid,
		&Greet{Name: "Bouhani"},
		time.Second,
	)

	logger.Info(reply.(*Greeting).Message)

	_ = system.Stop(ctx)
}
