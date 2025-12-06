package id

import (
	"github.com/google/uuid"
)

// Generate returns a new UUID string
func Generate() string {
	return uuid.New().String()
}

func IsValidUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)

}
