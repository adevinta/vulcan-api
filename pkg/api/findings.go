/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

type FindingOverride struct {
	ID             string    `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	UserID         string    `json:"user_id" validate:"required"`
	FindingID      string    `json:"finding_id" validate:"required"`
	StatusPrevious string    `json:"status_previous" validate:"required"`
	Status         string    `json:"status" validate:"required"`
	Notes          string    `json:"notes" validate:"required"`
	Tag            string    `json:"tag" validate:"required"`
	CreatedAt      time.Time `json:"-"`
}
