/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) FindTeamMember(teamID string, userID string) (*api.UserTeam, error) {
	teamMember := &api.UserTeam{TeamID: teamID, UserID: userID}

	findTeam := &api.Team{ID: teamID}
	res := db.Conn.Model(&api.Team{}).
		Find(&findTeam)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	res = db.Conn.
		Preload("User").
		Preload("Team").
		Find(&teamMember)

	if res.Error != nil {
		if !db.NotFoundError(res.Error) {
			return nil, db.logError(errors.Database(res.Error))
		}
	}

	return teamMember, nil
}

func (db vulcanitoStore) CreateTeamMember(teamMember api.UserTeam) (*api.UserTeam, error) {
	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	findUser := api.User{ID: teamMember.UserID}
	res := tx.Find(&findUser)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, db.logError(errors.Validation("User does not exist"))
	}

	findTeamMember := teamMember
	res = tx.Find(&findTeamMember)
	if res.RowsAffected > 0 {
		tx.Rollback()
		return nil, db.logError(errors.Duplicated("User is already a member of this team"))
	}

	if res.Error != nil {
		if !db.NotFoundError(res.Error) {
			tx.Rollback()
			return nil, db.logError(errors.Database(res.Error))
		}
	}

	teamMember.User = nil
	res = tx.Create(&teamMember)
	if res.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Create(res.Error))
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	db.Conn.Preload("Team").Preload("User").First(&teamMember)
	return &teamMember, nil
}

func (db vulcanitoStore) UpdateTeamMember(teamMember api.UserTeam) (*api.UserTeam, error) {
	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	findTeamMember := teamMember
	res := tx.Find(&findTeamMember)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, db.logError(errors.Validation("User is not a member of this team"))
	}

	if res.Error != nil {
		tx.Rollback()
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	res = tx.Omit("Team").Omit("User").Save(&teamMember)
	if res.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Update(res.Error))
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	db.Conn.Preload("Team").Preload("User").First(&teamMember)
	return &teamMember, nil
}

func (db vulcanitoStore) DeleteTeamMember(teamID string, userID string) error {
	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	teamMember := &api.UserTeam{TeamID: teamID, UserID: userID}
	res := tx.Find(&teamMember)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return db.logError(errors.Validation("User is not a member of this team"))
	}

	if res.Error != nil {
		tx.Rollback()
		if db.NotFoundError(res.Error) {
			return db.logError(errors.NotFound(res.Error))
		}
		return db.logError(errors.Database(res.Error))
	}

	res = tx.Delete(&teamMember)

	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	return nil
}
