/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/errors"
)

// UpdateRecipients ...
func (db vulcanitoStore) UpdateRecipients(teamID string, emails []string) error {
	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	result := tx.Delete(api.Recipient{}, "team_id = ?", teamID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	for _, e := range emails {
		r := &api.Recipient{
			TeamID: teamID,
			Email:  e,
		}

		result := tx.Create(r)
		if result.Error != nil {
			tx.Rollback()
			return db.logError(errors.Create(result.Error))
		}
	}

	// Commit the transaction
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	return nil
}

// ListRecipients ...
func (db vulcanitoStore) ListRecipients(teamID string) ([]*api.Recipient, error) {
	rs := []*api.Recipient{}
	result := db.Conn.Find(&rs, "team_id = ?", teamID)
	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return rs, nil
}
