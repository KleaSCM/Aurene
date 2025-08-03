/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: main.go
Description: Main entry point for the Aurene CPU scheduler. Initializes the scheduler
with signal handling for graceful shutdown and executes the CLI command structure.
*/

package main

import (
	"os"
	"os/signal"
	"syscall"

	"aurene/cmd"
	"aurene/internal/logger"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger := logger.New()
	logger.Info("🌌 Aurene - The Crystalline Scheduler starting up...")

	if err := cmd.Execute(); err != nil {
		logger.Error("Failed to execute command: %v", err)
		os.Exit(1)
	}

	<-sigChan
	logger.Info("Shutting down Aurene scheduler gracefully...")
}
