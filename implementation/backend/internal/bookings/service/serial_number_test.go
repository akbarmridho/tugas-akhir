package service

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"tugas-akhir/backend/internal/orders/entity"
)

func TestSerialNumberGenerator_Generate(t *testing.T) {
	t.Run("Successful Generation with Fixed Values", func(t *testing.T) {
		// Arrange
		orderItem := entity.OrderItem{
			OrderID:      12345,
			TicketSeatID: 67890,
		}

		// Create test generator with deterministic values
		generator := NewSerialNumberGenerator()

		// Override random reader with a fixed bytes reader
		fixedRandomBytes := []byte{1, 2, 3}
		generator.randomReader = bytes.NewReader(fixedRandomBytes)

		// Act
		serialNumber, err := generator.Generate(orderItem)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check format and components
		parts := strings.Split(serialNumber, "-")
		if len(parts) != 3 {
			t.Fatalf("Expected 3 parts separated by '-', got %d parts: %s", len(parts), serialNumber)
		}

		// Check prefix format
		expectedNumber := fmt.Sprintf("%03d%03d", orderItem.OrderID%1000, orderItem.TicketSeatID%1000)
		if parts[1] != expectedNumber {
			t.Errorf("Expected number to be %s, got: %s", expectedNumber, parts[1])
		}

		// Check random component
		if len(parts[2]) != 4 {
			t.Errorf("Expected random component to be 4 characters, got %d characters: %s", len(parts[2]), parts[2])
		}
	})

	t.Run("Large IDs are properly truncated", func(t *testing.T) {
		// Arrange
		orderItem := entity.OrderItem{
			OrderID:      1234567,    // This should become 567
			TicketSeatID: 9876543210, // This should become 210
		}

		// Create test generator with deterministic values
		generator := NewSerialNumberGenerator()

		// Override random reader with a fixed bytes reader
		fixedRandomBytes := []byte{4, 5, 6}
		generator.randomReader = bytes.NewReader(fixedRandomBytes)

		// Act
		serialNumber, err := generator.Generate(orderItem)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check format and components
		parts := strings.Split(serialNumber, "-")
		if len(parts) != 3 {
			t.Fatalf("Expected 3 parts separated by '-', got %d parts: %s", len(parts), serialNumber)
		}

		// Check prefix format
		expectedNumber := fmt.Sprintf("%03d%03d", orderItem.OrderID%1000, orderItem.TicketSeatID%1000)
		if parts[1] != expectedNumber {
			t.Errorf("Expected number to be %s, got: %s", expectedNumber, parts[1])
		}

		// Check random component
		if len(parts[2]) != 4 {
			t.Errorf("Expected random component to be 4 characters, got %d characters: %s", len(parts[2]), parts[2])
		}
	})

	t.Run("Small IDs are properly padded", func(t *testing.T) {
		// Arrange
		orderItem := entity.OrderItem{
			OrderID:      7,  // This should become 007
			TicketSeatID: 42, // This should become 042
		}

		// Create test generator with deterministic values
		generator := NewSerialNumberGenerator()

		// Override random reader with a fixed bytes reader
		fixedRandomBytes := []byte{4, 5, 6}
		generator.randomReader = bytes.NewReader(fixedRandomBytes)

		// Act
		serialNumber, err := generator.Generate(orderItem)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check format and components
		parts := strings.Split(serialNumber, "-")
		if len(parts) != 3 {
			t.Fatalf("Expected 3 parts separated by '-', got %d parts: %s", len(parts), serialNumber)
		}

		// Check prefix format
		expectedNumber := fmt.Sprintf("%03d%03d", orderItem.OrderID%1000, orderItem.TicketSeatID%1000)
		if parts[1] != expectedNumber {
			t.Errorf("Expected number to be %s, got: %s", expectedNumber, parts[1])
		}

		// Check random component
		if len(parts[2]) != 4 {
			t.Errorf("Expected random component to be 4 characters, got %d characters: %s", len(parts[2]), parts[2])
		}
	})
}
