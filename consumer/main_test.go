package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MockChannel mocks the RabbitMQ channel
type MockChannel struct {
	mock.Mock
}

// MockConnection mocks the RabbitMQ connection
type MockConnection struct {
	mock.Mock
}

// Mock methods for Channel
func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	argsMock := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return argsMock.Get(0).(amqp.Queue), argsMock.Error(1)
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	argsMock := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return argsMock.Get(0).(<-chan amqp.Delivery), argsMock.Error(1)
}

func (m *MockChannel) Close() error {
	return m.Called().Error(0)
}

// Mock methods for Connection
func (m *MockConnection) Channel() (*MockChannel, error) {
	argsMock := m.Called()
	return argsMock.Get(0).(*MockChannel), argsMock.Error(1)
}

func (m *MockConnection) Close() error {
	return m.Called().Error(0)
}

// TestConsumerFunction tests the main function logic with mocks
func TestConsumerFunction(t *testing.T) {
	mockConn := new(MockConnection)
	mockChannel := new(MockChannel)

	// Mock QueueDeclare and Consume
	mockChannel.On("QueueDeclare", "go-queue", false, false, false, false, nil).
		Return(amqp.Queue{Name: "go-queue"}, nil)

	mockMessages := make(chan amqp.Delivery)
	mockChannel.On("Consume", "go-queue", "", true, false, false, false, nil).
		Return(mockMessages, nil)

	// Mock connection and channel closing
	mockChannel.On("Close").Return(nil)
	mockConn.On("Channel").Return(mockChannel, nil)
	mockConn.On("Close").Return(nil)

	// Simulate message delivery
	go func() {
		mockMessages <- amqp.Delivery{Body: []byte("Hello World")}
		close(mockMessages)
	}()

	// Pass mock connection to the consumer function
	Consumer(mockConn)

	// Assert expectations
	mockChannel.AssertExpectations(t)
	mockConn.AssertExpectations(t)
}

