package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	rabbit "github.com/skobelina/rabbit-mq"
	"github.com/streadway/amqp"
)

type AddFile struct {
	NewFileName string
}

type FileService struct {
	queue   amqp.Queue
	channel *amqp.Channel
}

type JsonMessage struct {
	Message string `json:"message"`
}

func main() {
	srv := new(FileService)
	conn, err := amqp.Dial(rabbit.Config.AMQPConnectionURL)
	if err != nil {
		log.Fatal("Can't connect to AMQP", err)
	}
	defer conn.Close()

	srv.channel, err = conn.Channel()
	if err != nil {
		log.Fatal("Can't create a amqpChannel", err)
	}
	defer srv.channel.Close()

	srv.queue, err = srv.channel.QueueDeclare("add", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Could not declare `add` queue", err)
	}

	http.HandleFunc("/upload", srv.uploadFile)
	http.ListenAndServe(":8080", nil)
}

func (s *FileService) uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		handlerError("Error retrieving the file", err, w)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		handlerError("Error reading file", err, w)
		return
	}

	format, ok := getFormat(handler.Filename)
	if !ok {
		handlerError("Can't upload file", errors.New("Wrong format"), w)
		return
	}

	log.Printf("Uploaded File: %+v\n", handler.Filename)

	filename := strings.Replace(uuid.New().String(), "-", "", -1) + format
	temp, err := os.Create(path.Join("./storage", filename))
	if err != nil {
		handlerError("Error relacing file", err, w)
		return
	}
	defer temp.Close()

	err = ioutil.WriteFile(path.Join("./storage", filename), fileBytes, os.ModePerm)
	if err != nil {
		handlerError("Error writing file", err, w)
		return
	}

	addFile := AddFile{NewFileName: filename}
	body, err := json.Marshal(addFile)
	if err != nil {
		handlerError("Error encoding JSON", err, w)
		return
	}

	err = s.channel.Publish("", s.queue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})
	if err != nil {
		handlerError("Error publishing file", err, w)
		return
	}
	resp, _ := json.Marshal(&JsonMessage{"File upload succefully"})
	w.Write(resp)
	w.Header().Set("Content-Type", "application/json")
}

func handlerError(message string, err error, w http.ResponseWriter) {
	log.Printf("%s: %s", message, err.Error())
	body, _ := json.Marshal(&JsonMessage{message})
	w.Write(body)
	w.Header().Set("Content-Type", "application/json")
}

func getFormat(str string) (string, bool) {
	tokens := strings.Split(str, ".")
	switch tokens[len(tokens)-1] {
	case "jpg", "png", "jpeg":
		return "." + tokens[len(tokens)-1], true
	default:
		return "", false
	}
}
