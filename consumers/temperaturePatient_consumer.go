package consumers

import (
	"encoding/json"
	"file-consumer/data"
	"file-consumer/utils"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProcessTemperaturePatientMessages(token string, urlApi string, msgs <-chan amqp.Delivery) {
    for d := range msgs {
        log.Printf("Recibido mensaje de temperatura paciente: %s", d.Body)

        var rawData map[string]interface{}
        if err := json.Unmarshal(d.Body, &rawData); err != nil {
            log.Printf("❌ Error al parsear JSON: %s", err)
            continue
        }

        fmt.Printf("🔍 data from temperature patient: %v\n", rawData)

        // Validar cada campo antes de hacer type assertion
        measurementUnit, ok := rawData["measurementUnit"].(string)
        if !ok || measurementUnit == "" {
            log.Printf("❌ Error: measurementUnit no válido: %v", rawData["measurementUnit"])
            continue
        }

        nameSensor, ok := rawData["nameSensor"].(string)
        if !ok || nameSensor == "" {
            log.Printf("❌ Error: nameSensor no válido: %v", rawData["nameSensor"])
            continue
        }

        // AQUÍ está el problema más común - information puede ser nil
        informationRaw, exists := rawData["information"]
        if !exists || informationRaw == nil {
            log.Printf("❌ Error: information no existe o es nil")
            continue
        }
        
        information, ok := informationRaw.(float64)
        if !ok {
            log.Printf("❌ Error: information no es float64, es: %T = %v", informationRaw, informationRaw)
            continue
        }

        // Validar UserCivilIDUserCivil
        userCivilRaw, exists := rawData["UserCivilIDUserCivil"]
        if !exists || userCivilRaw == nil {
            log.Printf("❌ Error: UserCivilIDUserCivil no existe o es nil")
            continue
        }

        userCivilID, ok := userCivilRaw.(float64)
        if !ok {
            log.Printf("❌ Error: UserCivilIDUserCivil no es float64, es: %T = %v", userCivilRaw, userCivilRaw)
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
            log.Printf("❌ Error al crear JSON estandarizado: %s", err)
            continue
        }

        if err := utils.SendToAPI(urlApi, token, standardizedJSON); err != nil {
            log.Printf("❌ Error al enviar datos a la API: %s", err)
        } else {
            log.Printf("✅ Datos de temperatura paciente enviados exitosamente a la API: %s", standardizedJSON)
        }
    }
}