package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
	"tugas-akhir/backend/internal/orders/entity"
)

// SerialNumberGenerator handles generating unique serial numbers
type SerialNumberGenerator struct {
	randomReader io.Reader
}

// NewSerialNumberGenerator creates a new generator with defaults
func NewSerialNumberGenerator() *SerialNumberGenerator {
	return &SerialNumberGenerator{
		randomReader: rand.Reader,
	}
}

// Generate creates a serial number based on order item details
func (g *SerialNumberGenerator) Generate(item entity.OrderItem) (string, error) {
	// Generate random component (3 bytes = 4 chars in base64)
	randomBytes := make([]byte, 3)
	_, err := g.randomReader.Read(randomBytes)
	if err != nil {
		return "", errors.WithMessage(err, "failed to generate random bytes")
	}

	// Convert to base64 and clean up the string
	randomStr := base64.StdEncoding.EncodeToString(randomBytes)
	randomStr = strings.ReplaceAll(randomStr, "/", "X")
	randomStr = strings.ReplaceAll(randomStr, "+", "Y")
	randomStr = strings.TrimRight(randomStr, "=")

	// Combine all components into final serial number
	serialNumber := fmt.Sprintf("TIX-%03d%03d-%s", item.OrderID%1000, item.TicketSeatID%1000, randomStr[:4])

	return serialNumber, nil
}
