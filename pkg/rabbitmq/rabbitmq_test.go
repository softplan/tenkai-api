package rabbitmq

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func beforeTest() (RabbitImpl, amqp.Publishing, *amqp.Channel, *amqp.Connection) {
	rabbitImpl := RabbitImpl{}
	msg := amqp.Publishing{}
	channel := &amqp.Channel{}
	connection := &amqp.Connection{}
	return rabbitImpl, msg, channel, connection
}

func TestPublish(test *testing.T) {
	rabbitImpl, msg, channel, _ := beforeTest()
	assert.Panics(test, func() { rabbitImpl.Publish(channel, "", mock.Anything, false, false, msg) }, "Error on test Publish")
}

func TestGetConsumer(test *testing.T) {
	rabbitImpl, _, channel, _ := beforeTest()
	assert.Panics(test, func() { rabbitImpl.GetConsumer(channel, "", mock.Anything, false, false, false, false, amqp.Table{}) }, "Error on test GetConsumer")
}

func TestQueueDeclare(test *testing.T) {
	rabbitImpl, _, channel, _ := beforeTest()
	assert.Panics(test, func() { rabbitImpl.QueueDeclare(channel, mock.Anything, true, false, false, false, nil) }, "Error on test Publish")
}

func TestGetConnection(test *testing.T) {
	rabbitImpl, _, _, _ := beforeTest()
	assert.Panics(test, func() { rabbitImpl.GetConnection("") }, "Error on test GetConnection")
}

func TestGetChannel(test *testing.T) {
	rabbitImpl, _, _, connection := beforeTest()
	assert.Panics(test, func() { rabbitImpl.GetChannel(connection) }, "Error on test GetConnection")
}
