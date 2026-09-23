package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	goakt "github.com/tochemey/goakt/v4/actor"
	"github.com/tochemey/goakt/v4/discovery/static"
	"github.com/tochemey/goakt/v4/remote"
)

type EchoMessage struct {
	Text string
}

// EchoActor is a simple actor that logs received messages.
// It implements the full goakt.Actor interface (PreStart, Receive, PostStop).
type EchoActor struct{}

func (e *EchoActor) PreStart(*goakt.Context) error { return nil }

func (e *EchoActor) Receive(ctx *goakt.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case string:
		fmt.Printf("[%s] EchoActor received: %s\n", ctx.ActorSystem().Name(), msg)
		ctx.Response("ack: " + msg)
	case *EchoMessage:
		fmt.Printf("[%s] EchoActor received: %s\n", ctx.ActorSystem().Name(), msg.Text)
		ctx.Response(&EchoMessage{Text: "ack: " + msg.Text})
	default:
		ctx.Unhandled()
	}
}

func (e *EchoActor) PostStop(*goakt.Context) error { return nil }

func main() {
	// Determine which node to run based on args: "a" or "b"
	if len(os.Args) < 2 {
		log.Fatal("usage: go run main.go [a|b]")
	}
	nodeID := os.Args[1]

	ctx := context.Background()

	// --- Port assignments (unique per node) ---
	var (
		systemName    string
		remotingPort  int
		discoveryPort int
		peersPort     int
		seedAddr      string
	)

	switch nodeID {
	case "a":
		systemName = "system-a" // Node A system name
		remotingPort = 4041
		discoveryPort = 4001
		peersPort = 4002
		seedAddr = "127.0.0.1:4001" // Node A's own discovery port
	case "b":
		systemName = "system-a"
		remotingPort = 4051
		discoveryPort = 4011
		peersPort = 4012
		seedAddr = "127.0.0.1:4001" // Node B bootstraps from Node A
	default:
		log.Fatalf("unknown node: %s (use 'a' or 'b')", nodeID)
	}

	// --- Remoting config (required for clustering) ---
	remoteCfg := remote.NewConfig(
		"127.0.0.1",
		remotingPort,
		remote.WithSerializables(new(EchoMessage)),
	)

	// --- Static discovery: list of seed nodes (host:discoveryPort) ---
	discoveryCfg := static.Config{
		Hosts: []string{seedAddr},
	}
	provider := static.NewDiscovery(&discoveryCfg)

	// --- Cluster config ---
	clusterCfg := goakt.NewClusterConfig().
		WithDiscovery(provider).
		WithDiscoveryPort(discoveryPort).
		WithPeersPort(peersPort).
		WithKinds(&EchoActor{}) // Register the actor kind for cluster-wide placement

	// --- Create and start the actor system ---
	// The system name becomes the cluster identity internally.
	// Both nodes must have compatible cluster configs (same discovery mechanism).
	system, err := goakt.NewActorSystem(
		systemName,
		goakt.WithRemote(remoteCfg),
		goakt.WithCluster(clusterCfg),
	)
	if err != nil {
		log.Fatalf("failed to create actor system: %v", err)
	}

	if err := system.Start(ctx); err != nil {
		log.Fatalf("failed to start actor system: %v", err)
	}
	defer system.Stop(ctx)

	fmt.Printf("[%s] Actor system started (remoting=:%d, discovery=:%d)\n", systemName, remotingPort, discoveryPort)

	// --- Node A: spawn the actor and wait ---
	if nodeID == "a" {
		pid, err := system.Spawn(ctx, "echo", &EchoActor{})
		if err != nil {
			log.Fatalf("failed to spawn actor: %v", err)
		}
		fmt.Printf("[%s] Spawned 'echo': %s\n", systemName, pid.Path())

		// Keep Node A alive
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		return
	}

	// --- Node B: wait for cluster convergence, then message the remote actor ---
	fmt.Println("[B] Waiting for cluster to converge...")
	time.Sleep(5 * time.Second)

	// Resolve the actor by name (location-transparent lookup)
	pid, err := system.ActorOf(ctx, "echo")
	if err != nil {
		log.Fatalf("[B] failed to find 'echo': %v", err)
	}

	fmt.Printf("[B] Resolved 'echo' PID: %s\n", pid.Path())
	fmt.Printf("[B] IsRemote: %v\n", pid.IsRemote())

	// Send a message across the network
	resp, err := goakt.Ask(ctx, pid, &EchoMessage{Text: "hello from node B"}, 3*time.Second)
	if err != nil {
		log.Fatalf("[B] Ask failed: %v", err)
	}
	fmt.Printf("[B] Got response: %v\n", resp)

	fmt.Println("[B] Cross-node messaging succeeded.")
}
