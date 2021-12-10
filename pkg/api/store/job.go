/*
Copyright 2021 Adevinta
*/

package store

import (
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// FindJob retrieves a Job by its ID
func (db vulcanitoStore) FindJob(jobID string) (*api.Job, error) {
	if jobID == "" {
		return nil, db.logError(errors.Validation(`ID is empty`))
	}
	job := &api.Job{ID: jobID}
	res := db.Conn.Find(job)
	if res.Error != nil {
		if strings.HasPrefix(res.Error.Error(), `pq: invalid input syntax for type uuid`) {
			return nil, db.logError(errors.Validation(`ID is malformed`))
		}
		if !db.NotFoundError(res.Error) {
			return nil, db.logError(errors.Database(res.Error))
		}
	}

	if res.RowsAffected == 0 {
		return nil, db.logError(errors.NotFound("Job does not exists"))
	}

	return job, nil
}
