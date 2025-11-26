package model

import "time"

type Delivery struct {
	ID         int64
	CourierID  int64
	OrderID    string
	AssignedAt time.Time
	Deadline   time.Time
}
