package utils

import (
	"math/rand"
	"time"
)

// caracteres permitidos en el folio
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// función para inicializar la semilla del generador aleatorio
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateFolio genera un folio aleatorio de 8 caracteres
func GenerateFolio() string {
	folioLength := 8
	folio := make([]byte, folioLength)
	for i := range folio {
		folio[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(folio)
}
