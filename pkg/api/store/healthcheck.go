/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
)

// Healthcheck simply checks for database connectivity
func (db vulcanitoStore) Healthcheck() error {
	response := db.Conn.Exec("select 1;")
	if response.Error != nil {
		return errors.Database(response.Error)
	}

	return nil
}
