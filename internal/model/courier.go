package model

import (
	"regexp"
	"strings"
)

type Courier struct {
	ID     int64
	Name   string
	Phone  string
	Status string
}

func (c *Courier) ValidateData() bool {
	return strings.TrimSpace(c.Name) != "" &&
		strings.TrimSpace(c.Status) != "" &&
		regexp.MustCompile(`^\+?[1-9][0-9]{7,14}$`).MatchString(c.Phone)
}

func (c *Courier) Validate() bool {
	return c.ID > 0 && c.ValidateData()
}
