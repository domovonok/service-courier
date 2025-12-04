package factory

import "time"

const (
	TransportOnFoot  = "on_foot"
	TransportScooter = "scooter"
	TransportCar     = "car"
)

type DeliveryTimeCalculator interface {
	CalculateDeadline(fromTime time.Time) time.Time
}

type OnFootCalculator struct{}

func (c *OnFootCalculator) CalculateDeadline(fromTime time.Time) time.Time {
	return fromTime.Add(30 * time.Minute)
}

type ScooterCalculator struct{}

func (c *ScooterCalculator) CalculateDeadline(fromTime time.Time) time.Time {
	return fromTime.Add(15 * time.Minute)
}

type CarCalculator struct{}

func (c *CarCalculator) CalculateDeadline(fromTime time.Time) time.Time {
	return fromTime.Add(5 * time.Minute)
}

type DeliveryTimeCalculatorFactory interface {
	CreateCalculator(transportType string) DeliveryTimeCalculator
}

type deliveryTimeCalculatorFactory struct{}

func NewDeliveryTimeCalculatorFactory() DeliveryTimeCalculatorFactory {
	return &deliveryTimeCalculatorFactory{}
}

func (f *deliveryTimeCalculatorFactory) CreateCalculator(transportType string) DeliveryTimeCalculator {
	switch transportType {
	case TransportOnFoot:
		return &OnFootCalculator{}
	case TransportScooter:
		return &ScooterCalculator{}
	case TransportCar:
		return &CarCalculator{}
	default:
		return &OnFootCalculator{}
	}
}
