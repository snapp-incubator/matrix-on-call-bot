package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Shift struct {
	ID        int
	RoomID    string
	Sender    string
	Holders   string
	StartTime time.Time
	EndTime   *time.Time
}

type ShiftRepo interface {
	Create(s *Shift) error
	Get(roomID string) ([]Shift, error)
	Update(s *Shift) error
	Active(RoomID string) ([]Shift, error)
}

type SQLShiftRepo struct {
	DB *gorm.DB
}

func (ss *SQLShiftRepo) Create(s *Shift) error {
	return ss.DB.Create(s).Error
}

func (ss *SQLShiftRepo) Update(s *Shift) error {
	return ss.DB.Model(s).Where("end_time is null").Update(&Shift{EndTime: s.EndTime}).Error
}

func (ss *SQLShiftRepo) Get(roomID string) ([]Shift, error) {
	var res []Shift

	err := ss.DB.Where("room_id = ?", roomID).Order("start_time ASC").Find(&res).Error

	return res, err
}

func (ss *SQLShiftRepo) Active(roomID string) ([]Shift, error) {
	var res []Shift

	err := ss.DB.Where("room_id = ? AND end_time is null", roomID).Find(&res).Error

	return res, err
}
