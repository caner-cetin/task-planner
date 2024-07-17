package models

import "time"

type Student struct {
	ID        uint      `json:"-"`
	Username  string    `gorm:"uniqueIndex:compositeUsernameIndex;index;not null" json:"username"`
	Password  string    `json:"-"` // always skip encoding this field
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
