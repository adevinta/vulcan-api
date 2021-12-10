/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// JobRequest defines the information required to retrieve a job.
type JobRequest struct {
	ID string `json:"job_id" urlvar:"job_id"`
}

func makeFindJobEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*JobRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		job := api.Job{
			ID:        requestBody.ID,
			TeamID:    "TEAM_ID",
			Operation: "OPERATION",
			Status:    "STATUS",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		return Ok{job.ToResponse()}, nil
	}
}
