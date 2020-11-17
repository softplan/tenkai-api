package rabbitmq

import (
	"github.com/streadway/amqp"
)

//RabbitInterface interface
type RabbitInterface interface {
	GetConnection(uri string) *amqp.Connection
	GetChannel(conn *amqp.Connection) *amqp.Channel
	Publish(channel *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	GetConsumer(channel *amqp.Channel, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	QueueDeclare(channel *amqp.Channel, name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error)
}

//RabbitImpl struct
type RabbitImpl struct {
}

//Queues
const (
	InstallQueue       = "InstallQueue"
	ResultInstallQueue = "ResultInstallQueue"
	RepositoriesQueue  = "RepositoriesQueue"
	DeleteRepoQueue    = "DeleteRepoQueue"
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
func (rabbit RabbitImpl) GetChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		panic("Fail to open a channel with RabbitMQ Server")
	}
	return ch
}

//Publish a message on queue
func (rabbit RabbitImpl) Publish(channel *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return channel.Publish(exchange, key, mandatory, immediate, msg)
}

//GetConsumer queue
func (rabbit RabbitImpl) GetConsumer(channel *amqp.Channel, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

//QueueDeclare declare a queue
func (rabbit RabbitImpl) QueueDeclare(channel *amqp.Channel, name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return channel.QueueDeclare(name, true, false, false, false, nil)
}
