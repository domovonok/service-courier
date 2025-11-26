package factory

import "time"

const (
	TransportOnFoot  = "on_foot"
	TransportScooter = "scooter"
	TransportCar     = "car"
)

type DeliveryTimeFactory struct{}

func NewDeliveryTimeFactory() *DeliveryTimeFactory {
	return &DeliveryTimeFactory{}
}

func (f *DeliveryTimeFactory) CalculateDeadline(transportType string, fromTime time.Time) time.Time {
	switch transportType {
	case TransportOnFoot:
		return fromTime.Add(30 * time.Minute)
	case TransportScooter:
		return fromTime.Add(15 * time.Minute)
	case TransportCar:
		return fromTime.Add(5 * time.Minute)
	default:
		return fromTime.Add(30 * time.Minute)
	}
}
