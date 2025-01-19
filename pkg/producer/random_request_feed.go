package producer

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
)

// RandomRequestFeed is a request feed that generates random requests
type RandomRequestFeed struct {
	maxRequests  int
	requestCount int
}

// NewRandomRequestFeed creates a new RandomRequestFeed instance
func NewRandomRequestFeed(maxRequests int) *RandomRequestFeed {
	return &RandomRequestFeed{
		maxRequests:  maxRequests,
		requestCount: 0,
	}
}

// Next returns the next random request
func (r *RandomRequestFeed) Next() (models.RequestWrapper, error) {
	if r.maxRequests > 0 {
		r.requestCount++
	}

	types := []models.RequestType{models.AddItem, models.DeleteItem, models.GetItem, models.GetAll}
	cmdType := types[rand.Intn(len(types))]

	switch cmdType {
	case models.AddItem:
		payload := models.AddItemRequest{
			Key:   fmt.Sprintf("key%d", rand.Intn(1000)),
			Value: fmt.Sprintf("value%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.RequestWrapper{
			Type:    models.AddItem,
			Payload: rawPayload,
		}, nil

	case models.DeleteItem:
		payload := models.DeleteItemRequest{
			Key: fmt.Sprintf("key%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.RequestWrapper{
			Type:    models.DeleteItem,
			Payload: rawPayload,
		}, nil

	case models.GetItem:
		payload := models.GetItemRequest{
			Key: fmt.Sprintf("key%d", rand.Intn(1000)),
		}
		rawPayload, _ := json.Marshal(payload)
		return models.RequestWrapper{
			Type:    models.GetItem,
			Payload: rawPayload,
		}, nil

	case models.GetAll:
		rawPayload, _ := json.Marshal(models.GetAllItemsRequest{})
		return models.RequestWrapper{
			Type:    models.GetAll,
			Payload: rawPayload,
		}, nil

	default:
		return models.RequestWrapper{}, fmt.Errorf("unknown request type")
	}
}

// IsEmpty returns true if there are no more requests to generate
func (r *RandomRequestFeed) IsEmpty() bool {
	return r.maxRequests != -1 && r.requestCount >= r.maxRequests
}

// Close does nothing
func (r *RandomRequestFeed) Close() {}
