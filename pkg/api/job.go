/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"
)

// Job contains the status information of an asynchronous operation.
//
// In case of non-global operations it also contains the team ID associated to
// the operation.
type Job struct {
	ID        string `gorm:"primary_key:true"`
	TeamID    string `gorm:"Column:team_id"`
	Operation string
	Status    string `validate:"required"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j Job) ToResponse() *JobResponse {
	return &JobResponse{
		ID:        j.ID,
		TeamID:    j.TeamID,
		Operation: j.Operation,
		Status:    j.Status,
	}
}

type JobResponse struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
}
