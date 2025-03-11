package models

import "time"

type Message struct {
	ID        uint `gorm:"primaryKey"`
	From      string
	To        string
	Subject   string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
