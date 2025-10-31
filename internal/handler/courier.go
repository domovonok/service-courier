package handler

import (
	"regexp"
	"strings"
)

type courier struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

func (c *courier) validate() bool {
	return c.ID > 0 &&
		strings.TrimSpace(c.Name) != "" &&
		strings.TrimSpace(c.Status) != "" &&
		regexp.MustCompile(`^\+?[1-9][0-9]{7,14}$`).MatchString(c.Phone)
}
