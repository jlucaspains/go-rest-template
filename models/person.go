package models

import "time"

type Person struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required,min=3,max=100"`
	Email     string    `json:"email" binding:"required,min=3,max=100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
