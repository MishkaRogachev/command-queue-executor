package models

import (
	"encoding/json"
	"fmt"
)

// Command Types
type CommandType string

const (
	AddItem    CommandType = "addItem"
	DeleteItem CommandType = "deleteItem"
	GetItem    CommandType = "getItem"
	GetAll     CommandType = "getAllItems"
)

// CommandWrapper encapsulates all commands.
type CommandWrapper struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Request and Response Models

// AddItem
type AddItemRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AddItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteItem
type DeleteItemRequest struct {
	Key string `json:"key"`
}

type DeleteItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetItem
type GetItemRequest struct {
	Key string `json:"key"`
}

type GetItemResponse struct {
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetAllItems
type GetAllItemsRequest struct{}

type GetAllItemsResponse struct {
	Items   []KeyValuePair `json:"items"`
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func SerializeCommand(commandType CommandType, payload interface{}) (string, error) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	command := CommandWrapper{
		Type:    commandType,
		Payload: payloadData,
	}

	commandData, err := json.Marshal(command)
	if err != nil {
		return "", fmt.Errorf("failed to serialize command wrapper: %w", err)
	}

	return string(commandData), nil
}

func DeserializeCommandWrapper(raw string) (CommandWrapper, error) {
	var wrapper CommandWrapper
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return CommandWrapper{}, fmt.Errorf("failed to deserialize command wrapper: %w", err)
	}
	return wrapper, nil
}

func DeserializeCommand(raw string, target interface{}) (CommandType, error) {
	wrapper, err := DeserializeCommandWrapper(raw)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(wrapper.Payload, target); err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	return wrapper.Type, nil
}

func SerializeResponse(response interface{}) (string, error) {
	responseData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to serialize response: %w", err)
	}
	return string(responseData), nil
}

func DeserializeResponse(raw string, target interface{}) error {
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("failed to deserialize response: %w", err)
	}
	return nil
}
