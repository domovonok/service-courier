package handler

import "github.com/Avito-courses/course-go-avito-domovonok/internal/model"

type courierDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

func toDTO(c *model.Courier) courierDTO {
	return courierDTO{
		ID:     c.ID,
		Name:   c.Name,
		Phone:  c.Phone,
		Status: c.Status,
	}
}

func fromDTO(dto *courierDTO) *model.Courier {
	return &model.Courier{
		ID:     dto.ID,
		Name:   dto.Name,
		Phone:  dto.Phone,
		Status: dto.Status,
	}
}
