package consumers

import (
	"encoding/json"
	"file-consumer/data"
	"file-consumer/utils"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProcessAlcoholMessages(token string, urlApi string,msgs <-chan amqp.Delivery) {
    for d := range msgs {
        log.Printf("Recibido mensaje de humedad: %s", d.Body)

        var rawData map[string]interface{}
        if err := json.Unmarshal(d.Body, &rawData); err != nil {
            log.Printf("Error al parsear JSON: %s", err)
            continue
        }

        temperature := data.SensorData{
        }

        standardizedJSON, err := json.Marshal(temperature)
        if err != nil {
            log.Printf("Error al crear JSON estandarizado: %s", err)
            continue
        }

        if err := utils.SendToAPI(token,urlApi ,standardizedJSON); err != nil {
            log.Printf("Error al enviar datos a la API: %s", err)
        } else {
            log.Printf("Datos de humedad enviados exitosamente a la API: %s", standardizedJSON)
        }
    }
}