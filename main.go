package main

import (
	"context"
	"time"

	goakt "github.com/tochemey/goakt/v4/actor"
	"github.com/tochemey/goakt/v4/log"

	"github.com/hamdiBouhani/goakt-actor-examples/actors/bank"
	"github.com/hamdiBouhani/goakt-actor-examples/messages"
)

func main() {
	ctx := context.Background()
	logger := log.DefaultLogger

	system, err := goakt.NewActorSystem(
		"bank-system",
		goakt.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	err = system.Start(ctx)
	if err != nil {
		panic(err)
	}

	accountPID, err := system.Spawn(
		ctx,
		"account-1",
		&bank.Account{},
	)
	if err != nil {
		panic(err)
	}

	// deposit
	_, err = goakt.Ask(
		ctx,
		accountPID,
		&messages.Deposit{Amount: 100},
		time.Second,
	)
	if err != nil {
		panic(err)
	}

	// withdraw
	reply, err := goakt.Ask(
		ctx,
		accountPID,
		&messages.Withdraw{Amount: 40},
		time.Second)
	if err != nil {
		panic(err)
	}

	switch r := reply.(type) {
	case *messages.Balance:
		logger.Info("balance after withdraw", "amount", r.Amount)
	case *messages.InsufficientFunds:
		logger.Error("insufficient funds", "current", r.Current, "requested", r.Requested)
	}

	_ = system.Stop(ctx)
}
