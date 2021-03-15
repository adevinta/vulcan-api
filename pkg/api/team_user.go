/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

// UserTeam ...
type UserTeam struct {
	UserID    string    `gorm:"primary_key;AUTO_INCREMENT" json:"user_id" validate:"required"`
	User      *User     `json:"user" validate:"-"`
	TeamID    string    `gorm:"primary_key;AUTO_INCREMENT" json:"team_id" validate:"required"`
	Team      *Team     `json:"team" validate:"-"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (UserTeam) TableName() string {
	return "user_team"
}

func (ut UserTeam) ToResponse() *MemberResponse {
	response := &MemberResponse{}
	if ut.User != nil {
		response.User = *ut.User.ToResponse()
	}
	response.Role = ut.Role
	return response
}

type TeamMembersReponse struct {
	Team    *TeamResponse    `json:"team"`
	Members []MemberResponse `json:"members"`
}

type MemberResponse struct {
	User UserResponse `json:"user"`
	Role Role         `json:"role"`
}
