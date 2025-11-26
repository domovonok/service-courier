package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCourier_ValidateData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		courier  Courier
		expected bool
	}{
		{
			name: "valid courier with on_foot transport",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: true,
		},
		{
			name: "valid courier with scooter transport",
			courier: Courier{
				Name:          "Jane Smith",
				Phone:         "+9876543210",
				Status:        "busy",
				TransportType: "scooter",
			},
			expected: true,
		},
		{
			name: "valid courier with car transport",
			courier: Courier{
				Name:          "Bob Brown",
				Phone:         "1234567890",
				Status:        "available",
				TransportType: "car",
			},
			expected: true,
		},
		{
			name: "invalid - empty name",
			courier: Courier{
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - whitespace only name",
			courier: Courier{
				Name:          "   ",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - empty status",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (too short)",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "123",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (starts with 0)",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "0123456789",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (contains letters)",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "123abc7890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid transport type",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "bicycle",
			},
			expected: false,
		},
		{
			name: "invalid - empty transport type",
			courier: Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.courier.ValidateData()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCourier_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		courier  Courier
		expected bool
	}{
		{
			name: "valid courier with ID",
			courier: Courier{
				ID:            1,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: true,
		},
		{
			name: "invalid - zero ID",
			courier: Courier{
				ID:            0,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - negative ID",
			courier: Courier{
				ID:            -1,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - valid ID but invalid data",
			courier: Courier{
				ID:            1,
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.courier.Validate()
			assert.Equal(t, tt.expected, result)
		})
	}
}
