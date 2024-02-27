package domain

import (
	"github.com/google/uuid"
)

type Client struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Fullname    string    `json:"fullname"`
	YearOfBirth string    `json:"yearofbirth"`
	Limit       int       `json:"limit"`
	Balance     int       `json:"balance"`
}
