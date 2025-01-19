package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/MishkaRogachev/command-queue-executor/pkg/producer"
)

// Config holds the configuration settings for the client application
type Config struct {
	TimeoutMs          int    `json:"timeout_ms"`
	FeedType           string `json:"feed_type"`              // "file" or "random"
	CommandFile        string `json:"command_file,omitempty"` // For file feed
	RandomMax          int    `json:"random_max,omitempty"`   // For random feed
	RoutingKey         string `json:"routing_key"`
	MaxPendingRequests int    `json:"max_pending_requests"`
}

// loadConfig loads the configuration from a file
func loadConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func main() {
	configPath := flag.String("config", "cmd/client/config_file.json", "Path to the configuration file")
	flag.Parse()

	// Load the configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

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

	// Configure the producer based on feed type
	timeout := time.Duration(config.TimeoutMs) * time.Millisecond
	var prod *producer.Producer

	switch config.FeedType {
	case "file":
		if config.CommandFile == "" {
			log.Fatalf("CommandFile must be specified for file feed")
		}
		fileFeed, err := producer.NewFileRequestFeed(config.CommandFile)
		if err != nil {
			log.Fatalf("Failed to create file feed: %v", err)
		}
		defer fileFeed.Close()

		prod = producer.NewProducer(client, responseHandlerDebug, fileFeed, timeout, config.MaxPendingRequests)

	case "random":
		randomFeed := producer.NewRandomRequestFeed(config.RandomMax)
		prod = producer.NewProducer(client, responseHandlerDebug, randomFeed, timeout, config.MaxPendingRequests)

	default:
		log.Fatalf("Invalid FeedType: %s. Must be 'file' or 'random'.", config.FeedType)
	}

	// Start the producer
	prod.Start()
	defer prod.Close()

	log.Println("Producer is running. Press Ctrl+C to exit.")
	select {} // Keep the application running
}
