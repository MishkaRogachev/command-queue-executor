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

type Producer struct {
	client  mq.ClientMQ
	timeout time.Duration
}

func NewProducer(client mq.ClientMQ, timeout time.Duration) *Producer {
	return &Producer{
		client:  client,
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

	for scanner.Scan() {
		line := scanner.Text()

		var cmd models.CommandWrapper
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			return fmt.Errorf("failed to parse command: %w", err)
		}

		wg.Add(1)
		go func(command models.CommandWrapper) {
			defer wg.Done()

			rawCommand, err := json.Marshal(command)
			if err != nil {
				fmt.Printf("Error serializing command: %v\n", err)
				return
			}

			fmt.Println("Sending command:", string(rawCommand))
			responseChan, err := p.client.Request(string(rawCommand))
			if err != nil {
				fmt.Printf("Error sending command: %v\n", err)
				return
			}

			select {
			case response := <-responseChan:
				fmt.Printf("Received response: %s\n", response)
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
