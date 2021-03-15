/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

type Team struct {
	ID          string      `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Tag         string      `json:"tag" validate:"required"`
	CreatedAt   *time.Time  `json:"-"`
	UpdatedAt   *time.Time  `json:"-"`
	Assets      []*Asset    `json:"assets"`    // This line is infered from other tables.
	UserTeam    []*UserTeam `json:"user_team"` // This line is infered from other tables.
	Groups      []*Group
}

func (t Team) ToResponse() *TeamResponse {
	response := &TeamResponse{}
	response.ID = t.ID
	response.Name = t.Name
	response.Description = t.Description
	response.Tag = t.Tag

	return response
}

type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
}
