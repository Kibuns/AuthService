package main

import (
	"encoding/json"
	"log"

	"github.com/Kibuns/AuthService/DAL"
	"github.com/Kibuns/AuthService/Models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func receiveCredentials() {
    conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/") //locally change rabbitmq to localhost
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "user", // name
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    failOnError(err, "Failed to register a consumer")

    var forever chan struct{}

    go func() {
        for d := range msgs {
            var user Models.DetailedUser
            err := json.Unmarshal(d.Body, &user)
            failOnError(err, "Error deserializing message body")

            log.Printf("Received a message: %+v", user)
            DAL.RegisterUser(user)
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}

func receiveDeleted(){
    conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/") //locally change rabbitmq to localhost
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    deleteQueue, err := ch.QueueDeclare(
        "delete_credentials", // name
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

    deleteMsgs, err := ch.Consume(
        deleteQueue.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    failOnError(err, "Failed to register a consumer")

    var forever chan struct{}

    go func() {
        for t := range deleteMsgs {
            var username string
            err := json.Unmarshal(t.Body, &username)
            failOnError(err, "Error deserializing message body")
            log.Printf("Received a message to delete everything regarding user: %+v", username)
            DAL.DeleteAllOfUser(username)
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}