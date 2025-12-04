package factory_test

import (
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestDeliveryTimeCalculatorFactory_CreateCalculator(t *testing.T) {
	t.Parallel()

	f := factory.NewDeliveryTimeCalculatorFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		transportType string
		expected      time.Duration
	}{
		{
			name:          "on_foot transport",
			transportType: factory.TransportOnFoot,
			expected:      30 * time.Minute,
		},
		{
			name:          "scooter transport",
			transportType: factory.TransportScooter,
			expected:      15 * time.Minute,
		},
		{
			name:          "car transport",
			transportType: factory.TransportCar,
			expected:      5 * time.Minute,
		},
		{
			name:          "unknown transport defaults to on_foot",
			transportType: "bicycle",
			expected:      30 * time.Minute,
		},
		{
			name:          "empty transport defaults to on_foot",
			transportType: "",
			expected:      30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calculator := f.CreateCalculator(tt.transportType)
			result := calculator.CalculateDeadline(baseTime)
			expected := baseTime.Add(tt.expected)

			assert.Equal(t, expected, result)
		})
	}
}
