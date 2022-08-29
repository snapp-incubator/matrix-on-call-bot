package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type FollowUp struct {
	ID          int
	ShiftID     int
	Sender      string
	Initiator   string
	Description string
	Done        bool
	Category    string
	CreatedAt   time.Time
}

func (f FollowUp) TableName() string {
	return "follow_ups"
}

type FollowUpRepo interface {
	Create(f *FollowUp) error
	Get(ShiftID int) ([]FollowUp, error)
	Update(f *FollowUp) error
}

type SQLFollowUpRepo struct {
	DB *gorm.DB
}

func (fu *SQLFollowUpRepo) Create(f *FollowUp) error {
	return fu.DB.Create(f).Error
}

func (fu *SQLFollowUpRepo) Get(shiftID int) ([]FollowUp, error) {
	var res []FollowUp

	err := fu.DB.Where("shift_id = ?", shiftID).Find(&res).Error

	return res, err
}

func (fu *SQLFollowUpRepo) Update(f *FollowUp) error {
	return fu.DB.Model(f).Update(&FollowUp{Done: f.Done}).Error
}
