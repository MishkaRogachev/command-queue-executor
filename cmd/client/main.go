package main

import (
	"log"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/MishkaRogachev/command-queue-executor/pkg/producer"
)

// Config holds the configuration settings for the client application
type Config struct {
	Timeout     time.Duration `json:"timeout"`
	CommandFile string        `json:"command_file"`
	RoutingKey  string        `json:"routing_key"`
}

func loadConfig() Config {
	// Hardcoding config values for simplicity
	return Config{
		Timeout:     5 * time.Second,
		CommandFile: "test_data/test_commands_long.txt",
		RoutingKey:  "rpc_queue",
	}
}

func main() {
	config := loadConfig()

	// Initialize RabbitMQ client
	client, err := mq.NewClientRabbitMQ(mq.GetRabbitMQURL(), config.RoutingKey)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing RabbitMQ client: %v", err)
		}
	}()

	// Simple debug response handler
	responseHandlerDebug := func(response string) error {
		log.Printf("<< Received response: %s\n", response)
		return nil
	}

	prod := producer.NewFileProducer(client, responseHandlerDebug, config.Timeout)

	err = prod.ReadCommandsFromFile(config.CommandFile)
	if err != nil {
		log.Fatalf("Failed to process commands: %v", err)
	}

	log.Println("Done!")
}
