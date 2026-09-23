package bank

import (
	"github.com/hamdiBouhani/goakt-actor-examples/messages"
	goakt "github.com/tochemey/goakt/v4/actor"
)

type Account struct {
	balance int
}

func (a *Account) PreStart(*goakt.Context) error {
	return nil
}

func (a *Account) Receive(ctx *goakt.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *messages.Deposit:
		a.balance += msg.Amount

		// optional ack
		ctx.Response(
			&messages.Balance{
				Amount: a.balance,
			},
		)

	case *messages.Withdraw:
		if msg.Amount > a.balance {
			ctx.Response(
				&messages.InsufficientFunds{
					Current:   a.balance,
					Requested: msg.Amount,
				},
			)
			return
		}

		a.balance -= msg.Amount
		ctx.Response(
			&messages.Balance{
				Amount: a.balance,
			},
		)

	case *messages.GetBalance:
		ctx.Response(
			&messages.Balance{
				Amount: a.balance,
			},
		)

	default:
		ctx.Unhandled()
	}
}

func (a *Account) PostStop(*goakt.Context) error {
	return nil
}
