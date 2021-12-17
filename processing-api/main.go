package main

import (
	"encoding/json"

	"log"

	"github.com/disintegration/imaging"
	rabbit "github.com/skobelina/rabbit-mq"
	"github.com/streadway/amqp"
)

type AddFile struct {
	NewFileName string
}

func main() {
	conn, err := amqp.Dial(rabbit.Config.AMQPConnectionURL)
	if err != nil {
		log.Fatal("Can't connect to AMQP", err)
	}
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	if err != nil {
		log.Fatal("Can't create a amqpChannel", err)
	}
	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Could not declare `add` queue", err)
	}

	err = amqpChannel.Qos(1, 0, false)
	if err != nil {
		log.Fatal("Could not configure QoS", err)
	}

	fileChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Could not register consumer", err)
	}

	for d := range fileChannel {
		log.Printf("Received a file: %s", d.Body)
		addFile := &AddFile{}
		err := json.Unmarshal(d.Body, addFile)
		if err != nil {
			log.Printf("Error decoding JSON: %s", err)
			continue
		}
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging file: %s", err)
			continue
		} else {
			log.Printf("Acknowledged file")
		}
		compression(addFile.NewFileName)
	}
}

func compression(filename string) error {
	file, err := imaging.Open("./storage/" + filename)
	if err != nil {
		log.Println(err)
		return err
	}
	// Resize image from 1000 to 400 while preserving the aspect ration
	// Supported resize filters: NearestNeighbor, Box, Linear, Hermite, MitchellNetravali,
	// CatmullRom, BSpline, Gaussian, Lanczos, Hann, Hamming, Blackman, Bartlett, Welch, Cosine.
	dst := imaging.Resize(file, 400, 0, imaging.Linear)
	err = imaging.Save(dst, ("./storage/" + filename))
	if err != nil {
		log.Println(err)
	}
	return err
}
