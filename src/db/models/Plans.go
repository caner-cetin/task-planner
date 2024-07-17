package models

import (
	"database/sql/driver"
	"time"
)

type Plan struct {
	ID                uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title             string     `json:"title"`
	Description       *string    `json:"description"`
	Status            PlanStatus `sql:"type:ENUM('DRAFT', 'TO_DO', 'IN_PROGRESS', 'DONE')" gorm:"column:plan_status" json:"status"`
	EstimatedDeadline *time.Time `json:"estimated_deadline"`
	Time              time.Time  `json:"time"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	StudentID         uint       `json:"-"`
	Student           Student    `json:"-"`
}

type PlanStatus int64

const (
	DRAFT       PlanStatus = 0
	TO_DO       PlanStatus = 1
	IN_PROGRESS PlanStatus = 2
	DONE        PlanStatus = 3
)

func (self *PlanStatus) Scan(value interface{}) error { *self = PlanStatus(value.(int64)); return nil }
func (self PlanStatus) Value() (driver.Value, error)  { return int64(self), nil }
