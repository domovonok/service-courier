package handler

import (
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type courierDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	TransportType string `json:"transport_type"`
}

func toDTO(c *model.Courier) courierDTO {
	return courierDTO{
		ID:            c.ID,
		Name:          c.Name,
		Phone:         c.Phone,
		Status:        c.Status,
		TransportType: c.TransportType,
	}
}

func fromDTO(dto *courierDTO) *model.Courier {
	return &model.Courier{
		ID:            dto.ID,
		Name:          dto.Name,
		Phone:         dto.Phone,
		Status:        dto.Status,
		TransportType: dto.TransportType,
	}
}

type assignRequestDTO struct {
	OrderID string `json:"order_id"`
}

type assignResponseDTO struct {
	CourierID        int64     `json:"courier_id"`
	OrderID          string    `json:"order_id"`
	TransportType    string    `json:"transport_type"`
	DeliveryDeadline time.Time `json:"delivery_deadline"`
}

type unassignRequestDTO struct {
	OrderID string `json:"order_id"`
}

type unassignResponseDTO struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	CourierID int64  `json:"courier_id"`
}
