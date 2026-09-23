package main

import (
	"context"
	"net"
	"time"

	goakt "github.com/tochemey/goakt/v4/actor"
	"github.com/tochemey/goakt/v4/discovery/selfmanaged"
	"github.com/tochemey/goakt/v4/log"
	"github.com/tochemey/goakt/v4/remote"

	"github.com/hamdiBouhani/goakt-actor-examples/actors/bank"
	"github.com/hamdiBouhani/goakt-actor-examples/messages"
)

func main() {
	ctx := context.Background()
	logger := log.DefaultLogger

	// 1. Remoting is a hard requirement for clustering.
	remoteConfig := remote.NewConfig("127.0.0.1", 4041)

	// 2. Self-managed discovery configured for loopback (localhost).
	// SelfAddress: How other nodes reach this node's discovery port.
	// BroadcastAddress: Loopback for local-only discovery.
	provider := selfmanaged.NewDiscovery(&selfmanaged.Config{
		ClusterName:      "bank-system",
		SelfAddress:      "127.0.0.1:7946",       // Discovery port on localhost
		BroadcastAddress: net.IPv4(127, 0, 0, 1), // Ensure loopback broadcast
	})

	// 3. Cluster configuration registers the actor kind and sets ports.
	clusterConfig := goakt.NewClusterConfig().
		WithDiscovery(provider).
		WithDiscoveryPort(4042).   // Gossip/membership traffic
		WithPeersPort(4043).       // Cluster registry replication
		WithKinds(&bank.Account{}) // Register actor kind for cluster placement

	// 4. Create the actor system with remoting and clustering enabled.
	system, err := goakt.NewActorSystem(
		"bank-system", // Must be the same across all nodes in the cluster
		goakt.WithLogger(logger),
		goakt.WithRemote(remoteConfig),
		goakt.WithCluster(clusterConfig),
	)
	if err != nil {
		panic(err)
	}

	if err = system.Start(ctx); err != nil {
		panic(err)
	}
	defer system.Stop(ctx)

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
