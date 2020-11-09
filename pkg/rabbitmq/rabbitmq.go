package rabbitmq

import (
	"github.com/streadway/amqp"
)

//RabbitInterface interface
type RabbitInterface interface {
	GetConnection(uri string) *amqp.Connection
	GetChannel() *amqp.Channel
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	GetConsumer(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

//RabbitImpl struct
type RabbitImpl struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

//Queues
const (
	InstallQueue       = "InstallQueue"
	ResultInstallQueue = "ResultInstallQueue"
)

//GetConnection to the RabbitMQ Server
func (rabbit RabbitImpl) GetConnection(uri string) *amqp.Connection {
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic("Fail to connect RabbitMQ Server")
	}
	return conn
}

//GetChannel with rabbitMQ Server
func (rabbit RabbitImpl) GetChannel() *amqp.Channel {
	ch, err := rabbit.Conn.Channel()
	if err != nil {
		panic("Fail to open a channel with RabbitMQ Server")
	}
	return ch
}

//Publish a message on queue
func (rabbit RabbitImpl) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return rabbit.Channel.Publish(exchange, key, mandatory, immediate, msg)
}

//GetConsumer queue
func (rabbit RabbitImpl) GetConsumer(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return rabbit.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}
