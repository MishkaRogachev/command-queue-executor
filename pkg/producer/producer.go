package producer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

type ResponseHandlerFunc func(string) error

// Producer responsible for sending request to the message queue from a file and promoting responses to a handler
type Producer struct {
	client  mq.ClientMQ
	handler ResponseHandlerFunc
	timeout time.Duration
}

func NewProducer(client mq.ClientMQ, handler ResponseHandlerFunc, timeout time.Duration) *Producer {
	return &Producer{
		client:  client,
		handler: handler,
		timeout: timeout,
	}
}

func (p *Producer) ReadCommandsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		var cmd models.CommandWrapper
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			fmt.Printf("Failed to parse command: %v, skipping line %d", err, lineCount)
			continue
		}

		wg.Add(1)
		// Start a routine to send the command and await the response
		go func(command models.CommandWrapper) {
			defer wg.Done()

			rawCommand, err := json.Marshal(command)
			if err != nil {
				fmt.Printf("Error serializing command: %v\n", err)
				return
			}

			responseChan, err := p.client.Request(string(rawCommand))
			if err != nil {
				fmt.Printf("Error sending command: %v\n", err)
				return
			}

			select {
			case response := <-responseChan:
				p.handler(response)
			case <-time.After(p.timeout):
				fmt.Println("No response received in time")
			}
		}(cmd)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	wg.Wait()
	return nil
}
