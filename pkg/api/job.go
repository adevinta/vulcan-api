/*
Copyright 2021 Adevinta
*/

package api

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// JobStatusPending defines the status of a pending Job.
	JobStatusPending JobStatus = "PENDING"
	// JobStatusRunning defines the status of a running Job.
	JobStatusRunning JobStatus = "RUNNING"
	// JobStatusDone defines the status of a done Job.
	JobStatusDone JobStatus = "DONE"
)

type JobStatus string

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
	Status JobStatus  `validate:"required"`
	Result *JobResult `gorm:"Column:result"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// JobResult represents the result of a job. Data and Error fields are
// unstructured JSON fields which content may vary per each operation.
type JobResult struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

// Scan scans value into Jsonb, implements sql.Scanner interface.
// This method is necessary for GORM to known how to receive/save it into the database.
// Reference: https://gorm.io/docs/data_types.html
func (j *JobResult) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, j)

}

// Value returns json value, implements driver.Valuer interface.
// This method is necessary for GORM to known how to receive/save it into the database.
// Reference: https://gorm.io/docs/data_types.html
func (j *JobResult) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (r *JobResult) toJobResultResponse() JobResultResponse {
	return JobResultResponse{
		Data:  string(r.Data),
		Error: r.Error,
	}
}

func (j Job) Validate() error {
	switch j.Status {
	case JobStatusPending:
	case JobStatusRunning:
	case JobStatusDone:
	default:
		return errors.New("valid status are PENDING, RUNNING or DONE")
	}
	if !json.Valid(j.Result.Data) {
		return errors.New("invalid result data JSON")
	}
	return nil
}

func (j Job) ToResponse() *JobResponse {
	res := &JobResponse{
		ID:        j.ID,
		TeamID:    j.TeamID,
		Operation: j.Operation,
		Status:    j.Status,
	}
	if j.Result != nil {
		res.Result = j.Result.toJobResultResponse()
	}
	return res
}

// JobResponse represents the data for a Job that is
// returned as a response to Job queries through the API.
type JobResponse struct {
	ID        string            `json:"id"`
	TeamID    string            `json:"team_id,omitempty"`
	Operation string            `json:"operation"`
	Status    JobStatus         `json:"status"`
	Result    JobResultResponse `json:"result"`
}

type JobResultResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

type JobsRunner struct {
	Client JobsClient
}

type JobsClient interface {
	MergeDiscoveredAssets(ctx context.Context, teamID string, assets []Asset, groupName string) error
	FindJob(ctx context.Context, jobID string) (*Job, error)
	UpdateJob(ctx context.Context, job Job) (*Job, error)
}
