package model

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID        string
	Sender    string
	CreatedAt time.Time
}

type RoomRepo interface {
	Create(r *Room) error
}

type SQLRoomRepo struct {
	DB *gorm.DB
}

func (sr *SQLRoomRepo) Create(r *Room) error {
	return sr.DB.Create(r).Error
}
