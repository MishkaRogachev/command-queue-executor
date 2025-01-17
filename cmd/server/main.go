package main

import (
	"log"

	"github.com/MishkaRogachev/command-queue-executor/pkg/consumer"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

type Config struct {
	RabbitMQURL string `json:"rabbitmq_url"`
}

func loadConfig() Config {
	// Hardcoding config values for simplicity
	return Config{
		RabbitMQURL: "amqp://guest:guest@localhost",
	}
}

func main() {
	config := loadConfig()

	// Initialize RabbitMQ server
	server, err := mq.NewServerRabbitMQ(
		config.RabbitMQURL,
	)
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
