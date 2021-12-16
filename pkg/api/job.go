/*
Copyright 2021 Adevinta
*/

package api

import (
	"encoding/json"
	"errors"
	"time"
)

// Job contains the status information of an asynchronous operation.
//
// In case of non-global operations it also contains the team ID associated to
// the operation.
type Job struct {
	ID        string `gorm:"primary_key:true"`
	TeamID    string `gorm:"Column:team_id"`
	Operation string `validate:"required"`
	// Status possible values are:
	// - PENDING
	// - RUNNING
	// - DONE
	Status string `validate:"required"`
	// Result is a marshaled JSON representation of a JobResult.
	Result string `gorm:"Column:result"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j Job) Validate() error {
	switch j.Status {
	case "PENDING":
	case "RUNNING":
	case "DONE":
	default:
		return errors.New("valid status are PENDING, RUNNING or DONE")
	}
	if !json.Valid([]byte(j.Result)) {
		return errors.New("invalid JSON")
	}
	return nil
}

func (j Job) ToResponse() *JobResponse {
	var res JobResult
	_ = json.Unmarshal([]byte(j.Result), &res)
	return &JobResponse{
		ID:        j.ID,
		TeamID:    j.TeamID,
		Operation: j.Operation,
		Status:    j.Status,
		Result:    res,
	}
}

type JobResponse struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id,omitempty"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	Result    JobResult `json:"result"`
}

type JobResult struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}
