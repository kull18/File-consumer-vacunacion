package main

import (
	"file-consumer/consumers"
	"file-consumer/utils"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    token, errToken := utils.LogConsumer()

    if errToken != nil {
        log.Fatalf("Error al generar token: %s", errToken)
    }

	urlRabbit := os.Getenv("URLRABBIT")
	
	conn, err := amqp.Dial(urlRabbit)
    if err != nil {
        log.Fatalf("Error al conectar a RabbitMQ: %s", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Error al abrir canal RabbitMQ: %s", err)
    }
    defer ch.Close()

    humidityMsgs, err := utils.SetupConsumer(ch, "")
    if err != nil {
        log.Fatalf("Error al configurar consumidor de humedad: %s", err)
    }

    alcoholMsgs, err := utils.SetupConsumer(ch, "")
    if err != nil {
        log.Fatalf("Error al configurar consumidor de temperatura: %s", err)
    }


    temperatureAmbientalMsgs, err := utils.SetupConsumer(ch, "")
    if err != nil {
        log.Fatalf("Error al configurar consumidor de luz: %s", err)
    }

	temperaturePatientMsgs, err := utils.SetupConsumer(ch, "")
    if err != nil {
        log.Fatalf("Error al configurar consumidor de luz: %s", err)
    }

	urlApi1 := os.Getenv("URL_API_1")

	urlApi2 := os.Getenv("URL_API_2")
  
    go consumers.ProcessAlcoholMessages(token, urlApi2,alcoholMsgs)
	go consumers.ProcessTemperaturePatientMessages(token, urlApi2,temperaturePatientMsgs)


    go consumers.ProcessHumidityMessages(token, urlApi1, humidityMsgs)
    go consumers.ProcessTemperatureAmbientalMessages(token, urlApi1 ,temperatureAmbientalMsgs)


    log.Println("Esperando mensajes. Presiona CTRL+C para salir.")
    var forever chan struct{}
    <-forever
}