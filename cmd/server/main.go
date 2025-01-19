package main

import (
	"log"

	"github.com/MishkaRogachev/command-queue-executor/pkg/consumer"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

// Config holds the configuration settings for the server application
type Config struct {
	RoutingKey string `json:"routing_key"`
}

func loadConfig() Config {
	// Hardcoding config values for simplicity
	return Config{
		RoutingKey: "rpc_queue",
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

	// Initialize RequestHandlerOrderedMap
	handler := consumer.NewRequestHandlerOrderedMap()

	// Start the server's request handler
	err = server.ServeHandler(handler.Execute)
	if err != nil {
		log.Fatalf("Failed to start server handler: %v", err)
	}

	log.Println("RabbitMQ server is running. Press Ctrl+C to exit...")
	select {} // Keep the server running
}
