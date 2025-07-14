package services

import (
	"testing"
)

func TestLuhnService_Validate(t *testing.T) {
	luhnService := NewLuhnService()

	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid Luhn number",
			number:   "12345678903",
			expected: true,
		},
		{
			name:     "Valid Luhn number 2",
			number:   "9278923470",
			expected: true,
		},
		{
			name:     "Invalid Luhn number",
			number:   "12345678904",
			expected: false,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: false,
		},
		{
			name:     "Non-numeric string",
			number:   "123abc456",
			expected: false,
		},
		{
			name:     "Single digit",
			number:   "5",
			expected: false,
		},
		{
			name:     "Valid short number",
			number:   "12345674",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := luhnService.Validate(tt.number)
			if result != tt.expected {
				t.Errorf("Validate(%s) = %v, expected %v", tt.number, result, tt.expected)
			}
		})
	}
}
