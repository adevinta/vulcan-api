/*
Copyright 2021 Adevinta
*/

package store

import (
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"gorm.io/gorm"
)

// FindJob retrieves a Job by its ID
func (db vulcanitoStore) FindJob(jobID string) (*api.Job, error) {
	if jobID == "" {
		return nil, db.logError(errors.Validation(`ID is empty`))
	}
	job := &api.Job{ID: jobID}
	res := db.Conn.Find(job)
	if res.Error != nil {
		if strings.HasPrefix(res.Error.Error(), `pq: invalid input syntax for type uuid (SQLSTATE 22P02)`) {
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
	res := tx.Create(&job)
	err := res.Error
	if err != nil {
		return nil, db.logError(errors.Create(err))
	}
	tx.First(&job)
	return &job, nil
}

func (db vulcanitoStore) UpdateJob(job api.Job) (*api.Job, error) {
	res := db.Conn.Model(&job).Updates(&job)
	err := res.Error
	if err != nil {
		return nil, db.logError(errors.Update(err))
	}
	db.Conn.First(&job)
	return &job, nil
}
