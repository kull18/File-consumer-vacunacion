package consumers

import (
	"encoding/json"
	"file-consumer/data"
	"file-consumer/utils"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProcessTemperatureHieleraMessages(token string, urlApi string, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		log.Printf("Recibido mensaje de humedad: %s", d.Body)

		var rawData map[string]interface{}
		if err := json.Unmarshal(d.Body, &rawData); err != nil {
			log.Printf("Error al parsear JSON: %s", err)
			continue
		}

		temperatureHielera := data.SensorDataVaccine{
			MeasurementUnit:  rawData["measurementUnit"].(string),
			NameSensor:       rawData["nameSensor"].(string),
			Information:      rawData["information"].(float64),
			IDVaccineBox:     int(rawData["idVaccineBox"].(float64)),
			IDSensorsVaccine: int(rawData["idSensorsVaccine"].(float64)),
		}

		standardizedJSON, err := json.Marshal(temperatureHielera)
		if err != nil {
			log.Printf("Error al crear JSON estandarizado: %s", err)
			continue
		}

		if err := utils.SendToAPI(urlApi,token, standardizedJSON); err != nil {
			log.Printf("Error al enviar datos a la API: %s", err)
		} else {
			log.Printf("Datos de humedad enviados exitosamente a la API: %s", standardizedJSON)
		}
	}
}
