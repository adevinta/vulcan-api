/*
Copyright 2021 Adevinta
*/

package store

import (
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/jinzhu/gorm"
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
		return nil, db.logError(errors.NotFound("Job does not exist"))
	}

	return job, nil
}

func (db vulcanitoStore) createJobTx(tx *gorm.DB, job api.Job) (*api.Job, error) {
	res := tx.Preload("Team").Create(&job)
	err := res.Error
	if err != nil {
		return nil, db.logError(errors.Create(err))
	}
	tx.Preload("Team").First(&job)
	return &job, nil
}

func (db vulcanitoStore) updateJob(job api.Job) (*api.Job, error) {
	res := db.Conn.Preload("Team").Update(&job)
	err := res.Error
	if err != nil {
		return nil, db.logError(errors.Update(err))
	}
	db.Conn.Preload("Team").First(&job)
	return &job, nil
}
