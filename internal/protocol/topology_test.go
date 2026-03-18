package protocol

import (
	"context"
	"errors"
	"strings"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TestConsumerValidateWithEmptyName tests validation of consumer with empty name
func TestConsumerValidateWithEmptyName(t *testing.T) {
	consumer := Consumer{
		Name: "",
		Exec: func(ctx context.Context, msg Message) error {
			return nil
		},
	}

	err := consumer.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "consumer name cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumerValidateWithNullHandler tests validation of consumer with null handler
func TestConsumerValidateWithNullHandler(t *testing.T) {
	consumer := Consumer{
		Name: "test-consumer",
		Exec: nil,
	}

	err := consumer.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "handler cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumerValidateWithValidData tests validation of consumer with valid data
func TestConsumerValidateWithValidData(t *testing.T) {
	consumer := Consumer{
		Name: "test-consumer",
		Exec: func(ctx context.Context, msg Message) error {
			return nil
		},
		Queue: Queue{
			Name:              "test-queue",
			ShouldCreateQueue: true,
		},
	}

	err := consumer.Validate()
	if err != nil {
		t.Errorf("Expected no validation error, but got: %v", err)
	}
}

// TestConsumerMsg tests Msg method
func TestConsumerMsg(t *testing.T) {
	consumer := Consumer{}
	msgChan := make(chan amqp.Delivery, 1)
	consumer.Msg(msgChan)

	if consumer.Listen() != msgChan {
		t.Error("Msg did not configure the channel correctly")
	}
}

// TestConsumerListen tests Listen method
func TestConsumerListen(t *testing.T) {
	consumer := Consumer{}
	msgChan := make(chan amqp.Delivery, 1)
	consumer.Msg(msgChan)

	if consumer.Listen() != msgChan {
		t.Error("Listen did not return the configured channel")
	}
}

// TestConsumerHandler tests Handler method
func TestConsumerHandler(t *testing.T) {
	consumer := Consumer{}
	handlerCalled := false
	handler := func(ctx context.Context, msg Message) error {
		handlerCalled = true
		return nil
	}

	consumer.Handler(handler)

	if consumer.Exec == nil {
		t.Error("Handler did not configure Exec")
	}

	// Test if handler was configured correctly
	err := consumer.Exec(context.Background(), Message{})
	if err != nil {
		t.Errorf("Handler returned unexpected error: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

// TestConsumerExecute tests Execute method
func TestConsumerExecute(t *testing.T) {
	handlerCalled := false
	consumer := Consumer{
		Exec: func(ctx context.Context, msg Message) error {
			handlerCalled = true
			return nil
		},
	}

	err := consumer.Execute(context.Background(), Message{})
	if err != nil {
		t.Errorf("Execute returned unexpected error: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called via Execute")
	}
}

// TestQueueValidateWithEmptyName tests validation of queue with empty name
func TestQueueValidateWithEmptyName(t *testing.T) {
	queue := Queue{
		Name:              "",
		ShouldCreateQueue: true,
	}

	err := queue.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "queue name cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestQueueValidateWithValidData tests validation of queue with valid data
func TestQueueValidateWithValidData(t *testing.T) {
	queue := Queue{
		Name:              "test-queue",
		ShouldCreateQueue: true,
	}

	err := queue.Validate()
	if err != nil {
		t.Errorf("Expected no validation error, but got: %v", err)
	}
}

// TestQueueValidateWithInvalidHeaders tests validation of queue with invalid headers
func TestQueueValidateWithInvalidHeaders(t *testing.T) {
	queue := Queue{
		Name:              "test-queue",
		ShouldCreateQueue: true,
		Headers: map[string]any{
			"key": nil,
		},
	}

	err := queue.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "header") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestExchangeValidateWithEmptyName tests validation of exchange with empty name
func TestExchangeValidateWithEmptyName(t *testing.T) {
	exchange := Exchange{
		Name:                 "",
		Kind:                 Direct,
		ShouldCreateExchange: true,
	}

	err := exchange.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "exchange name cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestExchangeValidateWithInvalidKind tests validation of exchange with invalid kind
func TestExchangeValidateWithInvalidKind(t *testing.T) {
	exchange := Exchange{
		Name:                 "test-exchange",
		Kind:                 "invalid-kind",
		ShouldCreateExchange: true,
	}

	err := exchange.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "invalid exchange kind") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestExchangeValidateWithValidData tests validation of exchange with valid data
func TestExchangeValidateWithValidData(t *testing.T) {
	exchange := Exchange{
		Name:                 "test-exchange",
		Kind:                 Direct,
		ShouldCreateExchange: true,
	}

	err := exchange.Validate()
	if err != nil {
		t.Errorf("Expected no validation error, but got: %v", err)
	}
}

// TestConsumerValidateWithInvalidQueue tests validation of consumer with invalid queue
func TestConsumerValidateWithInvalidQueue(t *testing.T) {
	consumer := Consumer{
		Name: "test-consumer",
		Exec: func(ctx context.Context, msg Message) error {
			return nil
		},
		Queue: Queue{
			Name:              "",
			ShouldCreateQueue: true,
		},
	}

	err := consumer.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "queue validation failed") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumerValidateWithInvalidExchange tests validation of consumer with invalid exchange
func TestConsumerValidateWithInvalidExchange(t *testing.T) {
	consumer := Consumer{
		Name: "test-consumer",
		Exec: func(ctx context.Context, msg Message) error {
			return nil
		},
		Queue: Queue{
			Name:              "test-queue",
			ShouldCreateQueue: true,
		},
		Exchange: &Exchange{
			Name:                 "",
			Kind:                 Direct,
			ShouldCreateExchange: true,
		},
	}

	err := consumer.Validate()
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "exchange validation failed") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumerExecuteWithError tests Execute method with handler error
func TestConsumerExecuteWithError(t *testing.T) {
	expectedErr := errors.New("handler error")
	consumer := Consumer{
		Exec: func(ctx context.Context, msg Message) error {
			return expectedErr
		},
	}

	err := consumer.Execute(context.Background(), Message{})
	if err == nil {
		t.Error("Expected handler error, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Unexpected error: %v", err)
	}
}
