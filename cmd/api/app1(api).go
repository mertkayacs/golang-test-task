package main

//The main api class for POST and RabbitMQ

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

//Helper error method for RabbitMQ
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

//Message struct to unMarshall
type Message struct {
	message  string `json:"message"`
	sender   string `json:"sender"`
	receiver string `json:"receiver"`
}

//BodyOfCall struct to unMarshall
type BodyOfCall struct {
	sender   string `json:"sender"`
	receiver string `json:"receiver"`
}

//Temp list of messages to publish into Redis and to read from in the third application
var messages = []Message{
	{message: "hi", sender: "Mert", receiver: "Catbyte"},
	{message: "hi mert", sender: "Catbyte", receiver: "Mert"},
	{message: "hi catbyte, I will send a test message next", sender: "Mert", receiver: "Catbyte"},
}

func main() {

	//Connection to rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	//GIN API PART RUNS ON localhost:8080
	r := gin.Default()

	//Get method to receive all of the messages (temp for part3)
	// r.GET("/messages", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, messages[0].message)
	// })

	//RETURNS MESSAGE LIST

	r.GET("/message/list", func(c *gin.Context) {

		var body BodyOfCall
		if err := c.ShouldBindJSON(&body); err != nil {
			//Sending bad request if it fails
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, "messages : "+body.sender)
	})

	//Post method to send a message (for part1)

	r.POST("/message", func(c *gin.Context) {
		var newMessage Message

		if err := c.ShouldBindJSON(&newMessage); err != nil {
			//Sending bad request if it fails
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		body := newMessage
		//byte_message, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Errorf("error", err.Error())
		}

		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("Message"),
				// Body:        []byte(byte_message),
			})
		failOnError(err, "Failed to publish a messages")
		log.Printf(" [x] Sent %s\n", body)

		messages = append(messages, newMessage)

		//Sending status OK if all goes well
		c.JSON(http.StatusOK, newMessage)
	})

	r.Run()
}
