package producer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/stretchr/testify/assert"
)

func mockServerHandler(msg string) string {
	cmdWrapper, err := models.DeserializeCommandWrapper(msg)
	fmt.Println(">> Request:", msg)
	if err != nil {
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: false,
			Message: "failed to parse command",
		})
		return response
	}

	switch cmdWrapper.Type {
	case models.AddItem:
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: true,
			Message: "item added",
		})
		return response

	case models.DeleteItem:
		response, _ := models.SerializeResponse(models.DeleteItemResponse{
			Success: true,
			Message: "item deleted",
		})
		return response

	case models.GetItem:
		response, _ := models.SerializeResponse(models.GetItemResponse{
			Success: true,
			Value:   "testValue1",
		})
		return response

	case models.GetAll:
		response, _ := models.SerializeResponse(models.GetAllItemsResponse{
			Success: true,
			Items: []models.KeyValuePair{
				{Key: "testKey1", Value: "testValue1"},
			},
		})
		return response

	default:
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: false,
			Message: "unknown command",
		})
		return response
	}
}

func responseHandlerDebug(response string) error {
	fmt.Printf("<< Response: %s\n", response)
	return nil
}

func TestProducerWithFileRequestFeed(t *testing.T) {
	// Create a new inproc server
	server := mq.NewInprocServer()
	defer server.Close()

	// Get the requests channel
	reqCh, err := server.ListenForRequests()
	assert.NoError(t, err)

	// Start a goroutine to read requests + reply via mockServerHandler
	go func() {
		for req := range reqCh {
			resp := mockServerHandler(req.Data)
			_ = server.Reply(req.CorrelationID, resp)
		}
	}()

	// Create a client linked to the same server
	client := mq.NewInprocClient(server)

	fileFeed, err := NewFileRequestFeed("../../test_data/test_commands_short.txt")
	assert.NoError(t, err)
	defer fileFeed.Close()

	producer := NewProducer(client, responseHandlerDebug, fileFeed, 1*time.Second, 10)

	producer.Start()
	producer.Close()

	// Ensure all requests from file feed were processed
	assert.True(t, fileFeed.IsEmpty())
}

func TestProducerWithRandomRequestFeed(t *testing.T) {
	// Create a new inproc server
	server := mq.NewInprocServer()
	defer server.Close()

	// Get the requests channel
	reqCh, err := server.ListenForRequests()
	assert.NoError(t, err)

	// Start a goroutine to read requests + reply via mockServerHandler
	go func() {
		for req := range reqCh {
			resp := mockServerHandler(req.Data)
			_ = server.Reply(req.CorrelationID, resp)
		}
	}()

	// Create a client linked to the same server
	client := mq.NewInprocClient(server)

	randomFeed := NewRandomRequestFeed(50)
	producer := NewProducer(client, responseHandlerDebug, randomFeed, 1*time.Second, 10)

	producer.Start()
	producer.Close()

	// Ensure all requests from random feed were processed
	assert.True(t, randomFeed.IsEmpty())
}

func TestConcurrentProducersWithMixedFeeds(t *testing.T) {
	// Create a new inproc server
	server := mq.NewInprocServer()
	defer server.Close()

	// Get the requests channel
	reqCh, err := server.ListenForRequests()
	assert.NoError(t, err)

	// Start a goroutine to read requests + reply via mockServerHandler
	go func() {
		for req := range reqCh {
			resp := mockServerHandler(req.Data)
			_ = server.Reply(req.CorrelationID, resp)
		}
	}()

	// Create a clients linked to the same server
	client1 := mq.NewInprocClient(server)
	client2 := mq.NewInprocClient(server)
	client3 := mq.NewInprocClient(server)

	// File Feed
	fileFeed1, err := NewFileRequestFeed("../../test_data/test_commands_medium.txt")
	assert.NoError(t, err)
	defer fileFeed1.Close()

	fileFeed2, err := NewFileRequestFeed("../../test_data/test_commands_short.txt")
	assert.NoError(t, err)
	defer fileFeed2.Close()

	// Random Feed
	randomFeed := NewRandomRequestFeed(100)

	// Producers
	producer1 := NewProducer(client1, responseHandlerDebug, fileFeed1, 1*time.Second, 10)
	producer2 := NewProducer(client2, responseHandlerDebug, fileFeed2, 1*time.Second, 10)
	producer3 := NewProducer(client3, responseHandlerDebug, randomFeed, 1*time.Second, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		producer1.Start()
		producer1.Close()
	}()

	go func() {
		defer wg.Done()
		producer2.Start()
		producer2.Close()
	}()

	go func() {
		defer wg.Done()
		producer3.Start()
		producer3.Close()
	}()

	wg.Wait()

	// Ensure all feeds are empty after processing
	assert.True(t, fileFeed1.IsEmpty())
	assert.True(t, fileFeed2.IsEmpty())
	assert.True(t, randomFeed.IsEmpty())
}
