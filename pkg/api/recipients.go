/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

type Recipient struct {
	TeamID    string    `json:"team_id" gorm:"primary_key"`
	Email     string    `json:"email" gorm:"primary_key"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type RecipientResponse struct {
	Email string `json:"email"`
}

func (r Recipient) ToResponse() *RecipientResponse {
	response := RecipientResponse{
		Email: r.Email,
	}
	return &response
}
