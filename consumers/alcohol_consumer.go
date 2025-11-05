package consumers

import (
	"encoding/json"
	"file-consumer/data"
	"file-consumer/utils"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProcessAlcoholMessages(token string, urlApi string, msgs <-chan amqp.Delivery) {
    for d := range msgs {
        log.Printf("Recibido mensaje de temperatura paciente: %s", d.Body)

        var rawData map[string]interface{}
        if err := json.Unmarshal(d.Body, &rawData); err != nil {
            log.Printf("❌ Error al parsear JSON: %s", err)
            continue
        }

        fmt.Printf("🔍 data from temperature patient: %v\n", rawData)

        // Validar y extraer campos con seguridad
        measurementUnit, ok := rawData["measurementUnit"].(string)
        if !ok {
            log.Printf("❌ Error: measurementUnit no es string o no existe")
            continue
        }

        nameSensor, ok := rawData["nameSensor"].(string)
        if !ok {
            log.Printf("❌ Error: nameSensor no es string o no existe")
            continue
        }

        information, ok := rawData["information"].(float64)
        if !ok {
            log.Printf("❌ Error: information no es float64 o no existe")
            continue
        }

        userCivilID, ok := rawData["UserCivilIDUserCivil"].(float64)
        if !ok {
            log.Printf("❌ Error: UserCivilIDUserCivil no es float64 o no existe")
            continue
        }

        temperaturePatient := data.SensorDataCheck{
            MeasurementUnit:      measurementUnit,
            NameSensor:           nameSensor,
            Information:          information,
            UserCivilIDUserCivil: int(userCivilID),
        }

        standardizedJSON, err := json.Marshal(temperaturePatient)
        if err != nil {
            log.Printf("Error al crear JSON estandarizado: %s", err)
            continue
        }

        if err := utils.SendToAPI(urlApi, token, standardizedJSON); err != nil {
            log.Printf("Error al enviar datos a la API: %s", err)
        } else {
            log.Printf("✅ Datos de temperatura paciente enviados exitosamente a la API: %s", standardizedJSON)
        }
    }
}