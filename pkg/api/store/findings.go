/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) CreateFindingOverwrite(findingOverwrite api.FindingOverwrite) error {
	// Begin transaction.
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	// create entry in finding_overwrite
	result := tx.Create(&findingOverwrite)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	err := db.pushToOutbox(tx, opFindingOverwrite, findingOverwrite)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction.
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	return nil

}

func (db vulcanitoStore) ListFindingOverwrites(findingID string) ([]*api.FindingOverwrite, error) {
	findingOverwrites := []*api.FindingOverwrite{}
	result := db.Conn.Preload("User").Find(&findingOverwrites, "finding_id = ?", findingID)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return findingOverwrites, nil
}
