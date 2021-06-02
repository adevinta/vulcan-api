/*
Copyright 2021 Adevinta
*/

package store

import (
	"strings"
	"time"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/saml"
)

func (db vulcanitoStore) ListUsers() ([]*api.User, error) {
	users := []*api.User{}

	// Retrieve all teams
	result := db.Conn.
		Find(&users)
	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return users, nil
}

// CreateUser inserts a new user in the database
func (db vulcanitoStore) CreateUser(user api.User) (*api.User, error) {
	findUser := api.User{}
	res := db.Conn.Find(&findUser, "lower(email) = ?", strings.ToLower(user.Email))
	if len(findUser.ID) > 0 {
		return nil, db.logError(errors.Duplicated("User already exists"))
	}

	if res.Error != nil {
		if !db.NotFoundError(res.Error) {
			return nil, db.logError(errors.Database(res.Error))
		}
	}
	user.Email = strings.ToLower(user.Email)
	res = db.Conn.Create(&user)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}
	return &user, nil
}

// UpdateUser modifies an existing user in the database
func (db vulcanitoStore) UpdateUser(user api.User) (*api.User, error) {
	findUser := user
	// Make UpdateUser transactional.
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	res := tx.Find(&findUser)

	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, db.logError(errors.NotFound("User does not exists"))
	}

	if res.Error != nil {
		tx.Rollback()
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	res = tx.Model(user).Update(&user)
	if res.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Update(res.Error))
	}

	// Reload updated user to return the entity properly updated with all fields.
	updatedUser := user
	res = tx.Find(&updatedUser)

	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, db.logError(errors.NotFound("User does not exists"))
	}

	// Commit db transaction
	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return &updatedUser, nil
}

// CreateUserIfNotExists ...
func (db vulcanitoStore) CreateUserIfNotExists(userData saml.UserData) error {
	// Empty user
	u := &api.User{}

	// TODO: There are race conditions here that we should fix, for instance, if
	// a user does not exist two goroutines could check for that in the next
	// lines of code and then try to create the user, in which case one of them
	// will fail with and unexpected error.

	// Search database for a user with the email that Okta returned to us
	res := db.Conn.Find(u, "lower(email) = ?", strings.ToLower(userData.Email))
	if res.Error != nil {
		if !db.NotFoundError(res.Error) {
			return db.logError(errors.Database(res.Error))
		}
	}
	now := time.Now()
	// If there are no records with that email, then insert a new user.
	if res.RowsAffected == 0 {
		u.Firstname = userData.FirstName
		u.Lastname = userData.LastName
		u.Email = strings.ToLower(userData.Email)
		u.LastLogin = &now
		// Begin db transaction.
		tx := db.Conn.Begin()
		if tx.Error != nil {
			return db.logError(errors.Database(tx.Error))
		}

		res = tx.Create(u)
		if res.Error != nil {
			tx.Rollback()
			return db.logError(errors.Create(res.Error))
		}

		// PTVUL-913
		// In order to make easier onboard process when a user is created
		// we will check if is set as recipient for a team. If is set as
		// recipient for a team, it will become member of that team.
		// Eventually, we may want to modify this behavior.
		r := []api.Recipient{}
		recipients := tx.Find(&r, "lower(email) = ?", strings.ToLower(userData.Email))

		teamsJoined := []string{}
		if recipients.RowsAffected > 0 {
			for _, recipient := range r {
				membership := &api.UserTeam{
					TeamID: recipient.TeamID,
					UserID: u.ID,
					Role:   "member",
				}
				res = tx.Create(&membership)
				teamsJoined = append(teamsJoined, recipient.TeamID)
				if res.Error != nil {
					tx.Rollback()
					return db.logError(errors.Create(res.Error))
				}
			}
		}

		// Commit db transaction
		if tx.Commit().Error != nil {
			return db.logError(errors.Database(tx.Error))
		}
		if len(teamsJoined) > 0 {
			db.logger.Log("user", u.Email, "userID", u.ID, "membership", strings.Join(teamsJoined, ",")) // nolint
		}

		return nil
	}

	// res.RowsAffected greater than zero, so we are checking if the attributes
	// have changed on Okta
	u.LastLogin = &now
	if u.Firstname != userData.FirstName ||
		u.Lastname != userData.LastName {
		u.Firstname = userData.FirstName
		u.Lastname = userData.LastName
	}
	res = db.Conn.Model(u).Update(u)
	if res.Error != nil {
		return db.logError(errors.Create(res.Error))
	}
	return nil
}

// FindUserByID query a user by his ID
func (db vulcanitoStore) FindUserByID(userID string) (*api.User, error) {
	if userID == "" {
		return nil, db.logError(errors.Validation(`ID is empty`))
	}
	user := &api.User{ID: userID}
	res := db.Conn.Find(user)
	if res.Error != nil {
		if strings.HasPrefix(res.Error.Error(), `pq: invalid input syntax for type uuid`) {
			return nil, db.logError(errors.Validation(`ID is malformed`))
		}
		if !db.NotFoundError(res.Error) {
			return nil, db.logError(errors.Database(res.Error))
		}
	}

	if res.RowsAffected == 0 {
		return nil, db.logError(errors.NotFound("User does not exists"))
	}

	return user, nil
}

// FindUserByEmail query a user by his email
func (db vulcanitoStore) FindUserByEmail(email string) (*api.User, error) {
	if email == "" {
		return nil, db.logError(errors.Validation(`Email is empty`))
	}

	user := &api.User{}
	res := db.Conn.Find(&user, "lower(email) = ?", strings.ToLower(email))
	if res.Error != nil {
		if !db.NotFoundError(res.Error) {
			return nil, db.logError(errors.Database(res.Error))
		}
	}

	if res.RowsAffected == 0 {
		return nil, db.logError(errors.NotFound("User does not exists"))
	}

	return user, nil
}

func (db vulcanitoStore) DeleteUserByID(userID string) error {
	if userID == "" {
		return db.logError(errors.Validation(`ID is empty`))
	}
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return errors.Database(tx.Error)
	}
	res := tx.Exec("DELETE FROM user_team WHERE user_id=?", userID)
	if res.Error != nil && !db.NotFoundError(res.Error) {
		// This is arguable in the sense that we could check for an error when
		// rollbacking the transaction. We are not doing that because even if a
		// transacction is explicitly rolledback postgres will automatically
		// rollback it after certain amount of time or when the connection is
		// close.

		tx.Rollback()
		return errors.Database(errors.Database(res.Error))
	}
	user := &api.User{ID: userID}
	res = tx.Delete(user)
	err := res.Error
	if err != nil {
		_ = db.logError(errors.Database(err)) // nolint
	}
	if res.RowsAffected == 0 {
		return db.logError(errors.NotFound("User does not exists"))
	}

	if err := tx.Commit().Error; err != nil {
		return db.logError(errors.Database(err))
	}
	return nil
}
