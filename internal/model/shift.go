package model

import (
	"time"

	"gorm.io/gorm"
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
	Report(RoomID string, from time.Time, to time.Time) ([]ShiftReport, error)
}

type SQLShiftRepo struct {
	DB *gorm.DB
}

func (ss *SQLShiftRepo) Create(s *Shift) error {
	return ss.DB.Create(s).Error
}

func (ss *SQLShiftRepo) Update(s *Shift) error {
	return ss.DB.Model(s).Where("end_time is null").Updates(&Shift{EndTime: s.EndTime}).Error
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

type ShiftReport struct {
	Holders   string
	StartTime time.Time
	EndTime   *time.Time
}

// nolint: varnamelen
func (ss *SQLShiftRepo) Report(roomID string, from time.Time, to time.Time) ([]ShiftReport, error) {
	var res []ShiftReport

	err := ss.DB.Table("shifts").
		Select("holders", "start_time", "end_time").
		Where("room_id", roomID).
		Where("((start_time < ?) AND (end_time >= ?) AND (end_time <= ?) ) OR "+
			"((start_time >= ?) AND (start_time <= ?) AND (end_time >= ?) AND (end_time <= ?) ) OR "+
			"((start_time >= ?) AND (start_time <= ?) AND (end_time > ?)) OR "+
			"(end_time IS NULL AND start_time < ?)",
			from, from, to,
			from, to, from, to,
			from, to, to,
			to,
		).
		Find(&res).
		Error

	return res, err
}
