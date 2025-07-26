package utils

import amqp "github.com/rabbitmq/amqp091-go"

func SetupConsumer(ch *amqp.Channel,  queueName string) (<-chan amqp.Delivery, error) {


    msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
    return msgs, err
}