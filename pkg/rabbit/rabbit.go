package rabbit

import (
	"github.com/streadway/amqp"
)

//Rabbit struct
type Rabbit struct {
	Connection *amqp.Connection
	Channel *amqp.Channel
}

//Queues
const (
	InstallQueue = "InstallQueue"
)

//Connect to the RabbitMQ Server
func (rabbit *Rabbit) Connect(uri string) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic("Fail to connect RabbitMQ Server")
	}

	ch, err := conn.Channel()
	if err != nil {
		panic("Fail to open channel with RabbitMQ Server")
	}

	ch.QueueDeclare("InstallQueue", true, false, false, false, nil)

	rabbit.Channel = ch
	rabbit.Connection = conn
}