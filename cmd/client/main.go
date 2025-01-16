package main

import (
	"log"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/MishkaRogachev/command-queue-executor/pkg/producer"
)

type Config struct {
	RabbitMQURL  string        `json:"rabbitmq_url"`
	Timeout      time.Duration `json:"timeout"`
	RetryCount   int           `json:"retry_count"`
	RetryBackoff time.Duration `json:"retry_backoff"`
	CommandFile  string        `json:"command_file"`
}

func loadConfig() Config {
	// Hardcoding config values forn simplicity
	return Config{
		RabbitMQURL:  "amqp://guest:guest@localhost",
		Timeout:      5 * time.Second,
		RetryCount:   3,
		RetryBackoff: 1 * time.Second,
		CommandFile:  "commands.json",
	}
}

func main() {
	config := loadConfig()

	// Initialize the RabbitMQ client
	client, err := mq.NewClientRabbitMQ(
		config.RabbitMQURL,
		config.Timeout,
		config.RetryCount,
		config.RetryBackoff,
	)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing RabbitMQ client: %v", err)
		}
	}()

	// Create a new producer
	prod := producer.NewProducer(client)

	// Read and process commands from the file specified in the config
	err = prod.ReadCommandsFromFile(config.CommandFile)
	if err != nil {
		log.Fatalf("Failed to process commands: %v", err)
	}

	log.Println("All commands processed successfully!")
}
