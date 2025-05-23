package utility

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseNumberString(s string) (int64, int64, error) {
	// Split the string by hyphen
	parts := strings.Split(s, "-")

	// Check if we have exactly 2 parts
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format: expected 'number-number', got '%s'", s)
	}

	// Parse first number
	num1, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid first number: %v", err)
	}

	// Parse second number
	num2, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid second number: %v", err)
	}

	return num1, num2, nil
}
