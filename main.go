package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	server "github.com/dipendra-mule/low-latency-game-server/game-server"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	gameServer := server.NewGameServer()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		gameServer.Stop()
		os.Exit(0)
	}()

	if err := gameServer.Start(*port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
