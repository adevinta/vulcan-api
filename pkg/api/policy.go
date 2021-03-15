/*
Copyright 2021 Adevinta
*/

package api

import "time"

type Policy struct {
	ID                     string                    `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	TeamID                 string                    `json:"team_id"`
	Team                   *Team                     `json:"team"` // This line is infered from other tables.
	Name                   string                    `json:"name" validate:"required"`
	ChecktypeSettings      []*ChecktypeSetting       `json:"checktype_settings"` // This line is infered from other tables.
	ProgramsGroupsPolicies []*ProgramsGroupsPolicies `json:"program_policies"`   // This line is infered from other tables.
	Description            *string                   `json:"description,omitempty"`
	CreatedAt              *time.Time                `json:"-"`
	UpdatedAt              *time.Time                `json:"-"`
}

func (Policy) TableName() string {
	return "policies"
}

type PolicyResponse struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	Description            *string `json:"description,omitempty"`
	CheckTypeSettingsCount int     `json:"settings_count"`
}

func (p Policy) ToResponse() *PolicyResponse {
	response := PolicyResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
	}
	response.CheckTypeSettingsCount = len(p.ChecktypeSettings)
	return &response
}
