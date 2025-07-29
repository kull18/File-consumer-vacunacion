package utils

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// genera un folio aleatorio
func randomFolio(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateFolio genera un folio único, verificando que no exista en la base de datos
func GenerateFolio(db *sql.DB) (string, error) {
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		folio := randomFolio(8)

		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM UserCivil WHERE fol = ?)"
		err := db.QueryRow(query, folio).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("error al verificar folio: %v", err)
		}

		if !exists {
			return folio, nil
		}
	}

	return "", fmt.Errorf("no se pudo generar un folio único tras %d intentos", maxAttempts)
}
