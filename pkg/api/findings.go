/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

type FindingOverwrite struct {
	ID             string    `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	UserID         string    `json:"user_id" validate:"required"`
	User           *User     `json:"user,omitempty"` // This line is infered from column name "user_id".
	FindingID      string    `json:"finding_id" validate:"required"`
	StatusPrevious string    `json:"status_previous" validate:"required"`
	Status         string    `json:"status" validate:"required"`
	Notes          string    `json:"notes" validate:"required"`
	Tag            string    `json:"tag" validate:"required"`
	CreatedAt      time.Time `json:"-"`
}

type FindingOverwriteResponse struct {
	ID             string    `json:"id"`
	User           string    `json:"user"`
	FindingID      string    `json:"finding_id"`
	StatusPrevious string    `json:"status_previous"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes"`
	Tag            string    `json:"tag"`
	CreatedAt      time.Time `json:"created_at"`
}

func (fr FindingOverwrite) ToResponse() FindingOverwriteResponse {
	output := FindingOverwriteResponse{}

	output.ID = fr.ID
	output.User = fr.User.Email
	output.FindingID = fr.FindingID
	output.StatusPrevious = fr.StatusPrevious
	output.Status = fr.Status
	output.Notes = fr.Notes
	output.Tag = fr.Tag
	output.CreatedAt = fr.CreatedAt

	return output
}
