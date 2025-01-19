package models

import (
	"encoding/json"
	"fmt"
)

// RequestType represents the type of command
type RequestType string

const (
	// AddItem command type
	AddItem RequestType = "addItem"
	// DeleteItem command type
	DeleteItem RequestType = "deleteItem"
	// GetItem command type
	GetItem RequestType = "getItem"
	// GetAll command type
	GetAll RequestType = "getAllItems"
)

// RequestWrapper encapsulates all commands.
type RequestWrapper struct {
	Type    RequestType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Request and Response Models

// AddItemRequest represents the request to add an item
type AddItemRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AddItemResponse represents the response to an AddItemRequest
type AddItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteItemRequest represents the request to delete an item
type DeleteItemRequest struct {
	Key string `json:"key"`
}

// DeleteItemResponse represents the response to a DeleteItemRequest
type DeleteItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetItemRequest represents the request to get an item
type GetItemRequest struct {
	Key string `json:"key"`
}

// GetItemResponse represents the response to a GetItemRequest
type GetItemResponse struct {
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetAllItemsRequest represents the request to get all items
type GetAllItemsRequest struct{}

// GetAllItemsResponse represents the response to a GetAllItemsRequest
type GetAllItemsResponse struct {
	Items   []KeyValuePair `json:"items"`
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
}

// KeyValuePair represents a key-value pair
type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SerializeRequest serializes a request with a given payload
func SerializeRequest(requestType RequestType, payload interface{}) (string, error) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	request := RequestWrapper{
		Type:    requestType,
		Payload: payloadData,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request wrapper: %w", err)
	}

	return string(requestData), nil
}

// DeserializeCommandWrapper deserializes a command wrapper
func DeserializeCommandWrapper(raw string) (RequestWrapper, error) {
	var wrapper RequestWrapper
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return RequestWrapper{}, fmt.Errorf("failed to deserialize command wrapper: %w", err)
	}
	return wrapper, nil
}

// DeserializeRequest deserializes a command
func DeserializeRequest(raw string, target interface{}) (RequestType, error) {
	wrapper, err := DeserializeCommandWrapper(raw)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(wrapper.Payload, target); err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	return wrapper.Type, nil
}

// SerializeResponse serializes a response
func SerializeResponse(response interface{}) (string, error) {
	responseData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to serialize response: %w", err)
	}
	return string(responseData), nil
}

// DeserializeResponse deserializes a response
func DeserializeResponse(raw string, target interface{}) error {
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("failed to deserialize response: %w", err)
	}
	return nil
}
