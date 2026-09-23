package messages

// commands
type Deposit struct {
	Amount int
}
type Withdraw struct {
	Amount int
}

// queries
type GetBalance struct{}
type Balance struct {
	Amount int
}

// failures
type InsufficientFunds struct {
	Current   int
	Requested int
}

type Open struct{}
type Close struct{}
