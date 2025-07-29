package utils

import amqp "github.com/rabbitmq/amqp091-go"

func SetupConsumer(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	// Declarar cola (no durable para pruebas, ajusta según necesidad)
	_, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete cuando no la usa nadie
		false, // exclusiva
		false, // no esperar
		nil,   // argumentos
	)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(
		queueName,
		"",    // consumer tag, vacío para generar uno aleatorio
		true,  // auto-ack
		false, // exclusive
		false, // no-local (deprecated)
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
