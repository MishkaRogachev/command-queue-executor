package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MishkaRogachev/command-queue-executor/pkg/consumer"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

// Config holds the configuration settings for the server application
type Config struct {
	RoutingKey string `json:"routing_key"`
	Workers    int    `json:"workers"`
}

func loadConfig() Config {
	// Hardcoding config values for simplicity
	return Config{
		RoutingKey: "rpc_queue",
		Workers:    5,
	}
}

func main() {
	config := loadConfig()

	// Initialize RabbitMQ server
	server, err := mq.NewServerRabbitMQ(mq.GetRabbitMQURL(), config.RoutingKey)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ server: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("Error closing RabbitMQ server: %v", err)
		}
	}()

	// Initialize your handler (e.g. RequestHandlerOrderedMap, which has an Execute method)
	handler := consumer.NewRequestHandlerOrderedMap()

	// Create a Consumer with N worker goroutines
	con := consumer.NewConsumer(server, config.Workers, handler.Execute)

	// Start the consumer
	if err := con.Start(); err != nil {
		log.Fatalf("Failed to start concurrent consumer: %v", err)
	}
	defer con.Stop()

	log.Println("RabbitMQ server is running. Press Ctrl+C to exit...")

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
