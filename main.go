package main

import (
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"file-consumer/consumers"
	"file-consumer/data"
	"file-consumer/utils"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type FrecuenciaData struct {
	Intervalos []string  `json:"intervalos"`
	Marcas     []float64 `json:"marcas"`
	FA         []int     `json:"fa"`
	FR         []float64 `json:"fr"`
	FAC        []int     `json:"fac"`
}

type Hour struct {
	Valor float64
	Hora  string
	Type  string
}

var (
	wsTempConn      *websocket.Conn
	wsTempMutex     sync.Mutex
	wsHumidityConn  *websocket.Conn
	wsHumMutex      sync.Mutex
	wsUserCivilConn *websocket.Conn
	wsUserCivilMutex sync.Mutex

	datos      []Hour
	datosMutex sync.Mutex
)

func connectWebSocket(path string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: path}
	log.Printf("Conectando a WebSocket %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func reconnectWebSocket(path string, mutex *sync.Mutex, conn **websocket.Conn) error {
	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		c, err := connectWebSocket(path)
		if err == nil {
			mutex.Lock()
			if *conn != nil {
				(*conn).Close()
			}
			*conn = c
			mutex.Unlock()
			log.Printf("Reconectado a WebSocket %s", path)
			return nil
		}
		log.Printf("Error reconectando WS %s (intento %d): %v", path, i+1, err)
		time.Sleep(time.Second * time.Duration(2<<i))
	}
	return fmt.Errorf("no se pudo reconectar a WebSocket %s tras %d intentos", path, maxAttempts)
}

func sendToWebSocket(conn **websocket.Conn, mutex *sync.Mutex, path string, message []byte) error {
	mutex.Lock()
	c := *conn
	mutex.Unlock()

	if c == nil {
		if err := reconnectWebSocket(path, mutex, conn); err != nil {
			return err
		}
		mutex.Lock()
		c = *conn
		mutex.Unlock()
	}

	err := c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Error enviando mensaje WS en %s: %v. Intentando reconectar...", path, err)
		if err := reconnectWebSocket(path, mutex, conn); err != nil {
			return err
		}
		mutex.Lock()
		c = *conn
		mutex.Unlock()
		return c.WriteMessage(websocket.TextMessage, message)
	}

	return nil
}

func frequencyTable(datos []Hour) map[string]FrecuenciaData {
	resultados := make(map[string]FrecuenciaData)
	if len(datos) == 0 {
		return resultados
	}

	tipos := map[string]bool{}
	for _, d := range datos {
		tipos[d.Type] = true
	}

	for tipo := range tipos {
		vals := []float64{}
		for _, d := range datos {
			if d.Type == tipo {
				vals = append(vals, d.Valor)
			}
		}

		n := len(vals)
		if n == 0 {
			continue
		}

		intervalos := make([]string, n)
		fa := make([]int, n)
		fr := make([]float64, n)
		fac := make([]int, n)
		marcas := make([]float64, n)

		total := float64(n)
		acumulada := 0

		for i, v := range vals {
			intervalos[i] = fmt.Sprintf("%.2f", v)
			marcas[i] = v
			fa[i] = 1
			fr[i] = 100.0 / total
			acumulada++
			fac[i] = acumulada
		}

		resultados[tipo] = FrecuenciaData{
			Intervalos: intervalos,
			FA:         fa,
			FR:         fr,
			FAC:        fac,
			Marcas:     marcas,
		}
	}

	return resultados
}

func startPeriodicSender(tipo string, conn **websocket.Conn, mutex *sync.Mutex, path string, intervalo time.Duration) {
	go func() {
		for {
			time.Sleep(intervalo)

			datosMutex.Lock()
			copia := make([]Hour, len(datos))
			copy(copia, datos)
			datosMutex.Unlock()

			frec := frequencyTable(copia)
			if data, ok := frec[tipo]; ok {
				jsonData, err := json.Marshal(map[string]FrecuenciaData{tipo: data})
				if err != nil {
					log.Println("Error serializando datos de frecuencia:", err)
					continue
				}

				if err := sendToWebSocket(conn, mutex, path, jsonData); err != nil {
					log.Printf("Error enviando datos de %s al WebSocket: %v", tipo, err)
				}
			}
		}
	}()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando archivo .env: %v", err)
	}

	token, errToken := utils.LogConsumer()
	if errToken != nil {
		log.Fatalf("Error al generar token: %s", errToken)
	}
	urlRabbit := os.Getenv("URLRabbitMq")
	log.Printf("URL RabbitMQ: %s", urlRabbit)

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

	humidityMsgs, err := utils.SetupConsumer(ch, "humidity")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de humedad: %s", err)
	}

	alcoholMsgs, err := utils.SetupConsumer(ch, "alcohol")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de alcohol: %s", err)
	}

	temperatureAmbientalMsgs, err := utils.SetupConsumer(ch, "tempAm")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de temperatura ambiental: %s", err)
	}

	temperaturePatientMsgs, err := utils.SetupConsumer(ch, "tempPat")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de temperatura paciente: %s", err)
	}

	temperatureHieleraMsgs, err := utils.SetupConsumer(ch, "tempTp")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de hielera: %s", err)
	}

	userCivilMsgs, err := utils.SetupConsumer(ch, "userCivil")
	if err != nil {
		log.Fatalf("Error al configurar consumidor de userCivil: %s", err)
	}

	urlApi1 := os.Getenv("URL_API_1")
	urlApi2 := os.Getenv("URL_API_2")

	go consumers.ProcessAlcoholMessages(token, urlApi2, alcoholMsgs)
	go consumers.ProcessTemperaturePatientMessages(token, urlApi2, temperaturePatientMsgs)
	go consumers.ProcessHumidityMessages(token, urlApi1, humidityMsgs)
	go consumers.ProcessTemperatureAmbientalMessages(token, urlApi1, temperatureAmbientalMsgs)
	go consumers.ProcessTemperatureHieleraMessages(token, urlApi1, temperatureHieleraMsgs)

    	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	passwordEscaped := url.QueryEscape(password)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, passwordEscaped, host, port, dbname)
	log.Println("DSN usado:", dsn)


    db, err := sql.Open("mysql", dsn) 
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	wsTempConn, err = connectWebSocket("/ws/temperature-stats")
	if err != nil {
		log.Fatalf("Error conectando a WebSocket temperatura: %v", err)
	}
	defer wsTempConn.Close()

	wsHumidityConn, err = connectWebSocket("/ws/humidity-stats")
	if err != nil {
		log.Fatalf("Error conectando a WebSocket humedad: %v", err)
	}
	defer wsHumidityConn.Close()

	wsUserCivilConn, err = connectWebSocket("/ws/usercivil-stats")
	if err != nil {
		log.Fatalf("Error conectando a WebSocket UserCivil: %v", err)
	}
	defer wsUserCivilConn.Close()

	rand.Seed(time.Now().UnixNano())

	go func() {
		for msg := range temperatureHieleraMsgs {
			var sensor data.SensorDataVaccine
			if err := json.Unmarshal(msg.Body, &sensor); err != nil {
				log.Println("Error al deserializar temperatura hielera:", err)
				continue
			}

			valor := float64(sensor.Information) + (rand.Float64()*2 - 1)
			datosMutex.Lock()
			datos = append(datos, Hour{
				Valor: valor,
				Hora:  time.Now().Format("15:04:05"),
				Type:  "temperature",
			})
			datosMutex.Unlock()
		}
	}()

	go func() {
		for msg := range humidityMsgs {
			var sensor data.SensorDataVaccine
			if err := json.Unmarshal(msg.Body, &sensor); err != nil {
				log.Println("Error al deserializar humedad:", err)
				continue
			}

			valor := float64(sensor.Information) + (rand.Float64()*2 - 1)
			datosMutex.Lock()
			datos = append(datos, Hour{
				Valor: valor,
				Hora:  time.Now().Format("15:04:05"),
				Type:  "humidity",
			})
			datosMutex.Unlock()
		}
	}()

	// Nuevo consumidor de userCivil
	go func() {
		for msg := range userCivilMsgs {
			var input data.UserCivil
			if err := json.Unmarshal(msg.Body, &input); err != nil {
				log.Println("Error al deserializar UserCivil:", err)
				continue
			}

			// Generar folio único
			folio, err := utils.GenerateFolio(db)
			if err != nil {
				log.Println("Error generando folio:", err)
				continue
			}

			userCivilFormat := data.UserCivilFormat{
				Fol:                 folio,
				CorporalTemperature: input.CorporalTemperature,
				AlcoholBreat:        input.AlcoholBreat,
			}

			jsonData, err := json.Marshal(userCivilFormat)
			if err != nil {
				log.Println("Error serializando UserCivil:", err)
				continue
			}

			if err := sendToWebSocket(&wsUserCivilConn, &wsUserCivilMutex, "/ws/usercivil-stats", jsonData); err != nil {
				log.Printf("Error enviando datos de UserCivil al WebSocket: %v", err)
			}
		}
	}()

	startPeriodicSender("temperature", &wsTempConn, &wsTempMutex, "/ws/temperature-stats", 10*time.Second)
	startPeriodicSender("humidity", &wsHumidityConn, &wsHumMutex, "/ws/humidity-stats", 10*time.Second)

	log.Println("Esperando mensajes. Presiona CTRL+C para salir.")
	select {}
}
