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
		Exec: func(ctx context.Context, msg amqp.Delivery) error {
			return nil
		},
	}

	err := consumer.Validate()
	if err == nil {
		t.Error("Esperado erro de validação, mas obteve nil")
	}

	if !strings.Contains(err.Error(), "consumer name cannot be empty") {
		t.Errorf("Erro inesperado: %v", err)
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
		t.Error("Esperado erro de validação, mas obteve nil")
	}

	if !strings.Contains(err.Error(), "handler cannot be empty") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestConsumerValidateWithValidData tests validation of consumer with valid data
func TestConsumerValidateWithValidData(t *testing.T) {
	consumer := Consumer{
		Name: "test-consumer",
		Exec: func(ctx context.Context, msg amqp.Delivery) error {
			return nil
		},
	}

	err := consumer.Validate()
	if err != nil {
		t.Errorf("Não esperado erro de validação, mas obteve: %v", err)
	}
}

// TestConsumerMsg tests Msg method
func TestConsumerMsg(t *testing.T) {
	consumer := Consumer{}
	msgChan := make(chan amqp.Delivery, 1)
	consumer.Msg(msgChan)

	if consumer.Listen() != msgChan {
		t.Error("Msg não configurou o canal corretamente")
	}
}

// TestConsumerListen tests Listen method
func TestConsumerListen(t *testing.T) {
	consumer := Consumer{}
	msgChan := make(chan amqp.Delivery, 1)
	consumer.Msg(msgChan)

	if consumer.Listen() != msgChan {
		t.Error("Listen não retornou o canal configurado")
	}
}

// TestConsumerHandler tests Handler method
func TestConsumerHandler(t *testing.T) {
	consumer := Consumer{}
	handlerCalled := false
	handler := func(ctx context.Context, msg amqp.Delivery) error {
		handlerCalled = true
		return nil
	}

	consumer.Handler(handler)

	if consumer.Exec == nil {
		t.Error("Handler não configurou o Exec")
	}

	// Testa se o handler foi configurado corretamente
	err := consumer.Exec(context.Background(), amqp.Delivery{})
	if err != nil {
		t.Errorf("Handler retornou erro inesperado: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler não foi chamado")
	}
}

// TestConsumerExecute tests Execute method
func TestConsumerExecute(t *testing.T) {
	handlerCalled := false
	consumer := Consumer{
		Exec: func(ctx context.Context, msg amqp.Delivery) error {
			handlerCalled = true
			return nil
		},
	}

	err := consumer.Execute(context.Background(), amqp.Delivery{})
	if err != nil {
		t.Errorf("Execute retornou erro inesperado: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler não foi chamado via Execute")
	}
}

// TestConsumerExecuteWithError tests Execute method with handler error
func TestConsumerExecuteWithError(t *testing.T) {
	expectedErr := errors.New("handler error")
	consumer := Consumer{
		Exec: func(ctx context.Context, msg amqp.Delivery) error {
			return expectedErr
		},
	}

	err := consumer.Execute(context.Background(), amqp.Delivery{})
	if err == nil {
		t.Error("Esperado erro do handler, mas obteve nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Erro inesperado: %v", err)
	}
}
