package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LogConsumer() (string, error) {
	err := godotenv.Load(".env") 
	if err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	username := os.Getenv("USERNAME_API")
	password := os.Getenv("PASSWORD_API")

	if username == "" || password == "" {
		return "", fmt.Errorf("USERNAME o PASSWORD no definidos en el archivo .env")
	}

	fmt.Printf("username password %s , %s\n", username, password)

	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("error al serializar los datos de login: %v", err)
	}

	apiURL := os.Getenv("URL_API_POST")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error al hacer la solicitud: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Estado de la respuesta: %s", resp.Status)

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Cuerpo de la respuesta: %s", string(body))

	token := resp.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no se encontró el token en la cabecera Authorization")
	}

	// Eliminar prefijo "Bearer " si está presente
	const bearerPrefix = "Bearer "
	if len(token) > len(bearerPrefix) && token[:len(bearerPrefix)] == bearerPrefix {
		token = token[len(bearerPrefix):]
	}

	log.Printf("Token obtenido: %s", token)

	return token, nil
}
