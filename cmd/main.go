package main

import (
	"fmt"
	"log"
	"os"

	"github.com/k8-proxy/k8-go-comm/pkg/minio"
	"github.com/k8-proxy/k8-go-comm/pkg/rabbitmq"
	"github.com/streadway/amqp"
)

var (
	ProcessingOutcomeExchange   = "processing-outcome-exchange"
	ProcessingOutcomeRoutingKey = "processing-outcome"
	ProcessingOutcomeQueueName  = "processing-outcome-queue"

	AdaptationOutcomeExchange   = "adaptation-exchange"
	AdaptationOutcomeRoutingKey = "adaptation-exchange"
	AdaptationOutcomeQueueName  = "amq.rabbitmq.reply-to"

	inputMount                     = os.Getenv("INPUT_MOUNT")
	adaptationRequestQueueHostname = os.Getenv("ADAPTATION_REQUEST_QUEUE_HOSTNAME")
	adaptationRequestQueuePort     = os.Getenv("ADAPTATION_REQUEST_QUEUE_PORT")
	messagebrokeruser              = os.Getenv("MESSAGE_BROKER_USER")
	messagebrokerpassword          = os.Getenv("MESSAGE_BROKER_PASSWORD")

	transactionStorePath = os.Getenv("TRANSACTION_STORE_PATH")

	publisher *amqp.Channel
)

func main() {

	// Get a connection
	connection, err := rabbitmq.NewInstance(adaptationRequestQueueHostname, adaptationRequestQueuePort, messagebrokeruser, messagebrokerpassword)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Initiate a publisher on processing exchange
	publisher, err = rabbitmq.NewQueuePublisher(connection, AdaptationOutcomeExchange)
	if err != nil {
		log.Fatalf("could not start publisher %s", err)
	}
	defer publisher.Close()

	// Start a consumer
	msgs, ch, err := rabbitmq.NewQueueConsumer(connection, ProcessingOutcomeQueueName, ProcessingOutcomeExchange, ProcessingOutcomeRoutingKey)
	if err != nil {
		log.Fatalf("could not start consumer %s", err)
	}
	defer ch.Close()

	forever := make(chan bool)

	// Consume
	go func() {
		for d := range msgs {
			err := processMessage(d)
			if err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func processMessage(d amqp.Delivery) error {

	/*
		if d.Headers["file-id"] == nil ||
			d.Headers["source-file-location"] == nil ||
			d.Headers["clean-file-presigned-url"] == nil ||
			d.Headers["rebuilt-file-location"] == nil {
			return fmt.Errorf("Headers value is nil")
		}*/

	fmt.Printf("%+v\n", d.Headers)

	fileID := ""
	outputFileLocation := ""
	cleanPresignedURL := ""
	reportFileName := "report.xml"

	if d.Headers["file-id"] != nil {
		log.Printf("file id is ok")
		fileID = d.Headers["file-id"].(string)
	}
	if d.Headers["rebuilt-file-location"] != nil {
		log.Printf("rebuilt-file-location is ok")
		outputFileLocation = d.Headers["rebuilt-file-location"].(string)
	}
	if d.Headers["clean-presigned-url"] != nil {
		log.Printf("clean-presigned-url is ok")
		cleanPresignedURL = d.Headers["clean-presigned-url"].(string)
	}

	log.Printf("Received a message for file: %s, clean presigned url %s outputFileLocation %s", fileID, cleanPresignedURL, outputFileLocation)

	// Download the file to output file location
	err := minio.DownloadObject(cleanPresignedURL, outputFileLocation)
	if err != nil {
		return err
	}
	if d.Headers["report-presigned-url"] != nil {
		reportPresignedURL := d.Headers["report-presigned-url"].(string)
		reportPath := fmt.Sprintf("%s/%s", transactionStorePath, fileID)

		if _, err := os.Stat(reportPath); os.IsNotExist(err) {
			os.MkdirAll(reportPath, 0777)
		}

		reportFileLocation := fmt.Sprintf("%s/%s", reportPath, reportFileName)

		log.Println("report file location ", reportFileLocation)

		err := minio.DownloadObject(reportPresignedURL, reportFileLocation)
		if err != nil {
			return err
		}
	}

	d.Headers["file-outcome"] = "replace"
	// Publish the details to Rabbit

	err = rabbitmq.PublishMessage(publisher, "", d.Headers["reply-to"].(string), d.Headers, []byte(""))

	if err != nil {
		return err
	}

	return nil
}
