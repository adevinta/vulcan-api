/*
Copyright 2024 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// ListIssues returns all issues
func (db vulcanitoStore) ListIssues() ([]*api.Issue, error) {
	issues := []*api.Issue{}

	// Retrieve all issues
	result := db.Conn.
		Find(&issues)
	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return issues, nil
}
