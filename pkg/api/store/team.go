/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/jinzhu/gorm"
)

// CreateTeam inserts a new team in the database and includes the current user
// as an owner
func (db vulcanitoStore) CreateTeam(team api.Team, ownerEmail string) (*api.Team, error) {
	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	// Insert the new team
	result := tx.Create(&team)
	if result.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Create(result.Error))
	}

	// Retrieve the current user
	owner := api.User{}
	result = tx.Find(&owner, "email = ?", ownerEmail)
	if result.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Database(result.Error))
	}

	// Assign current user and default owners to team
	owners := append(db.defaults["owners"], owner.ID)
	err := assignOwnersToTeam(tx, team.ID, owners)
	if err != nil {
		tx.Rollback()
		return nil, db.logError(errors.Database(err))
	}

	// Commit the transaction
	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return &team, nil
}

// UpdateTeam updates a team in the database. It verifies if the team exists
// before actually making the update
func (db vulcanitoStore) UpdateTeam(team api.Team) (*api.Team, error) {
	findTeam := team

	res := db.Conn.Find(&findTeam)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	res = db.Conn.Model(&team).Updates(&team)

	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}

	db.Conn.First(&team)
	return &team, nil
}

func (db vulcanitoStore) DeleteTeam(teamID string) error {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	deletedTeam := &api.Team{}
	res := tx.Raw("SELECT * FROM teams WHERE id = ? FOR UPDATE", teamID).
		Scan(&deletedTeam)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return db.logError(errors.NotFound(res.Error))
		}
		return db.logError(errors.Database(res.Error))
	}

	// We are not going to delete Scans and Reports
	// Scans belongs to Scans Engine and will be stored there.
	// Reports are linked to Scans and therefore should be preserved as well.

	// Delete ChecktypeSettings
	res = tx.Delete(api.ChecktypeSetting{}, "policy_id in (select p.id from policies p where p.team_id = ?)", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Delete policies
	res = tx.Delete(api.Policy{}, "team_id = ?", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Delete asset groups
	res = tx.Delete(api.AssetGroup{}, "group_id in (select g.id from groups g where g.team_id =?)", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Delete groups
	res = tx.Delete(api.Group{}, "team_id = ?", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Delete programs
	res = tx.Delete(api.Program{}, "team_id = ?", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Delete recipients
	res = tx.Delete(api.Recipient{}, "team_id = ?", teamID)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Push to outbox so distributed tx is processed
	err := db.pushToOutbox(tx, opDeleteTeam, *deletedTeam)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete assets
	err = db.deleteAllAssetsTX(tx, *deletedTeam)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete team
	res = tx.Delete(deletedTeam)
	if res.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(res.Error))
	}

	// Commit the transaction
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	return nil
}

// FindTeam returns a team filtering by id
func (db vulcanitoStore) FindTeam(teamID string) (*api.Team, error) {
	team := &api.Team{ID: teamID}
	res := db.Conn.Preload("UserTeam").Preload("UserTeam.User").Find(&team)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return team, nil
}

// FindTeamByName returns a team filtering by name
func (db vulcanitoStore) FindTeamByName(name string) (*api.Team, error) {
	team := &api.Team{}
	res := db.Conn.Preload("UserTeam").Preload("UserTeam.User").Find(&team, "name = ?", name)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return team, nil
}

// FindTeamByTag returns a team filtering by tag
func (db vulcanitoStore) FindTeamByTag(tag string) (*api.Team, error) {
	team := &api.Team{}
	res := db.Conn.Preload("UserTeam").Preload("UserTeam.User").Find(&team, "tag = ?", tag)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return team, nil
}

// FindTeamByProgram returns the team that the given Program belongs to.
func (db vulcanitoStore) FindTeamByProgram(programID string) (*api.Team, error) {
	program := &api.Program{ID: programID}
	result := db.Conn.
		Preload("Team").
		Find(&program)

	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return program.Team, nil
}

// ListTeams returns all teams
func (db vulcanitoStore) ListTeams() ([]*api.Team, error) {
	teams := []*api.Team{}

	// Retrieve all teams
	result := db.Conn.
		Find(&teams)
	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return teams, nil
}

func (db vulcanitoStore) FindTeamsByUser(userID string) ([]*api.Team, error) {
	teams := []*api.Team{}

	user := api.User{ID: userID}
	result := db.Conn.Find(&user)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	// Retrieve all teams
	result = db.Conn.
		Joins("join user_team on user_team.team_id = teams.id and user_team.user_id = ?", userID).
		Find(&teams)

	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return teams, nil
}

// FindTeamByIDForUser returns the membership information of a user in a team.
func (db vulcanitoStore) FindTeamByIDForUser(ID, userID string) (*api.UserTeam, error) {
	teamUser := &api.UserTeam{}
	res := db.Conn.Preload("User").Find(teamUser, "team_id = ? and user_id = ?", ID, userID)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, errors.NotFound(res.Error)
		}
		return nil, db.logError(errors.Database(res.Error))
	}
	return teamUser, nil
}

func assignOwnersToTeam(tx *gorm.DB, teamID string, owners []string) error {
	for _, userID := range owners {
		// Define the association between Team and Owner
		teamUser := api.UserTeam{
			TeamID: teamID,
			UserID: userID,
			Role:   api.Owner,
		}

		// Persist the association
		result := tx.Create(&teamUser)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}
