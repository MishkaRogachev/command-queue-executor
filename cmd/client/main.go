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
		CommandFile:  "test_data/test_commands_long.txt",
	}
}

func main() {
	config := loadConfig()

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

	responseHandlerDebug := func(response string) error {
		log.Printf("<< Received response: %s\n", response)
		return nil
	}

	// NOTE: using the same timeout for response await
	prod := producer.NewFileProducer(client, responseHandlerDebug, config.Timeout)

	err = prod.ReadCommandsFromFile(config.CommandFile)
	if err != nil {
		log.Fatalf("Failed to process commands: %v", err)
	}

	log.Println("Done!")
}
