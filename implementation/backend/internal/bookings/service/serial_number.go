package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"
	"tugas-akhir/backend/internal/orders/entity"
)

func GenerateSerialNumber(item entity.OrderItem) (string, error) {
	// Create prefix based on event details (first 2 chars of order ID and ticket seat ID)
	prefix := fmt.Sprintf("TIX-%03d-%03d", item.OrderID%1000, item.TicketSeatID%1000)

	// Add timestamp component (YYMMDDhhmm format)
	timestamp := time.Now().Format("0601021504")

	// Generate random component (6 bytes = 8 chars in base64)
	randomBytes := make([]byte, 6)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.WithMessage(err, "failed to generate random bytes")
	}

	// Convert to base64 and clean up the string
	randomStr := base64.StdEncoding.EncodeToString(randomBytes)
	randomStr = strings.ReplaceAll(randomStr, "/", "X")
	randomStr = strings.ReplaceAll(randomStr, "+", "Y")
	randomStr = strings.TrimRight(randomStr, "=")

	// Combine all components into final serial number
	serialNumber := fmt.Sprintf("%s-%s-%s", prefix, timestamp, randomStr[:8])

	return serialNumber, nil
}
