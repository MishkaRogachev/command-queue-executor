package producer

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

type RandomProducer struct {
	client  mq.ClientMQ
	handler ResponseHandlerFunc
	timeout time.Duration
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func NewRandomProducer(client mq.ClientMQ, handler ResponseHandlerFunc, timeout time.Duration) *RandomProducer {
	return &RandomProducer{
		client:  client,
		handler: handler,
		timeout: timeout,
		stopCh:  make(chan struct{}),
	}
}

// Start generates random commands indefinitely until Stop is called
func (rp *RandomProducer) Start() {
	rp.wg.Add(1)
	go func() {
		defer rp.wg.Done()
		for {
			select {
			case <-rp.stopCh:
				fmt.Println("Stopping RandomProducer...")
				return
			default:
				command := rp.generateRandomCommand()
				rp.sendCommand(command)
				time.Sleep(100 * time.Millisecond) // Control the generation rate
			}
		}
	}()
}

// Stop signals the producer to stop generating commands
func (rp *RandomProducer) Stop() {
	close(rp.stopCh)
	rp.wg.Wait()
}

func (rp *RandomProducer) sendCommand(command models.CommandWrapper) {
	rawCommand, err := json.Marshal(command)
	if err != nil {
		fmt.Printf("Error serializing command: %v\n", err)
		return
	}

	responseChan, err := rp.client.Request(string(rawCommand))
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		return
	}

	select {
	case response := <-responseChan:
		rp.handler(response)
	case <-time.After(rp.timeout):
		fmt.Println("No response received in time")
	}
}

func (rp *RandomProducer) generateRandomCommand() models.CommandWrapper {
	types := []models.CommandType{models.AddItem, models.DeleteItem, models.GetItem, models.GetAll}
	cmdType := types[rand.Intn(len(types))]

	switch cmdType {
	case models.AddItem:
		payload := models.AddItemRequest{
			Key:   fmt.Sprintf("key%d", rand.Intn(1000)),
			Value: fmt.Sprintf("value%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.CommandWrapper{
			Type:    models.AddItem,
			Payload: rawPayload,
		}

	case models.DeleteItem:
		payload := models.DeleteItemRequest{
			Key: fmt.Sprintf("key%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.CommandWrapper{
			Type:    models.DeleteItem,
			Payload: rawPayload,
		}

	case models.GetItem:
		payload := models.GetItemRequest{
			Key: fmt.Sprintf("key%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.CommandWrapper{
			Type:    models.GetItem,
			Payload: rawPayload,
		}

	case models.GetAll:
		rawPayload, _ := json.Marshal(models.GetAllItemsRequest{})
		return models.CommandWrapper{
			Type:    models.GetAll,
			Payload: rawPayload,
		}

	default:
		rawPayload, _ := json.Marshal(models.GetAllItemsRequest{})
		return models.CommandWrapper{
			Type:    models.GetAll,
			Payload: rawPayload,
		}
	}
}
