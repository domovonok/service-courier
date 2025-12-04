package model_test

import (
	"testing"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCourier_ValidateData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		courier  model.Courier
		expected bool
	}{
		{
			name: "valid courier with on_foot transport",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: true,
		},
		{
			name: "valid courier with scooter transport",
			courier: model.Courier{
				Name:          "Jane Smith",
				Phone:         "+9876543210",
				Status:        "busy",
				TransportType: "scooter",
			},
			expected: true,
		},
		{
			name: "valid courier with car transport",
			courier: model.Courier{
				Name:          "Bob Brown",
				Phone:         "1234567890",
				Status:        "available",
				TransportType: "car",
			},
			expected: true,
		},
		{
			name: "invalid - empty name",
			courier: model.Courier{
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - whitespace only name",
			courier: model.Courier{
				Name:          "   ",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - empty status",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (too short)",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "123",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (starts with 0)",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "0123456789",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid phone (contains letters)",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "123abc7890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expected: false,
		},
		{
			name: "invalid - invalid transport type",
			courier: model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "bicycle",
			},
			expected: false,
		},
		{
			name: "invalid - empty transport type",
			courier: model.Courier{
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
		courier  model.Courier
		expected bool
	}{
		{
			name: "valid courier with ID",
			courier: model.Courier{
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
			courier: model.Courier{
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
			courier: model.Courier{
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
			courier: model.Courier{
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
