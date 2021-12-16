package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type AddFile struct {
	FileName string
	IdFile   string
}

func main() {
	conn, err := amqp.Dial(rabbit.Config.AMQPConnectionURL)
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")
	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	handleError(err, "Could not declare `add` queue")

	err = amqpChannel.Qos(1, 0, false)
	handleError(err, "Could not configure QoS")

	fileChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	handleError(err, "Could not register consumer")

	stopChan := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range fileChannel {
			log.Printf("Received a file: %s", d.Body)

			addFile := &AddFile{}

			err := json.Unmarshal(d.Body, addFile)
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging file: %s", err)
			} else {
				log.Printf("Acknowledged file")
			}
		}
	}()
	<-stopChan
}
