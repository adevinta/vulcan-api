/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) CreatePolicy(policy api.Policy) (*api.Policy, error) {
	res := db.Conn.Create(&policy)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}
	db.Conn.
		Preload("Team").
		First(&policy)
	return &policy, nil
}

func (db vulcanitoStore) ListPolicies(teamID string) ([]*api.Policy, error) {
	policies := []*api.Policy{}
	res := db.Conn.Preload("Team").
		Preload("ChecktypeSettings").
		Preload("ProgramsGroupsPolicies").
		Preload("ProgramsGroupsPolicies.Program").
		Preload("ProgramsGroupsPolicies.Policy").
		Preload("ProgramsGroupsPolicies.Policy.Team").
		Preload("ProgramsGroupsPolicies.Policy.ChecktypeSettings").
		Find(&policies, "team_id = ?", teamID)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}
	return policies, nil
}

func (db vulcanitoStore) FindPolicy(policyID string) (*api.Policy, error) {
	policy := &api.Policy{ID: policyID}
	res := db.Conn.Preload("Team").Preload("ChecktypeSettings").Find(&policy)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return policy, nil
}

func (db vulcanitoStore) UpdatePolicy(policy api.Policy) (*api.Policy, error) {
	result := db.Conn.Model(&policy).
		Preload("Group").
		Where("team_id = ?", policy.TeamID).
		Updates(policy)
	if result.Error != nil {
		return nil, db.logError(errors.Update(result.Error))
	}

	db.Conn.Preload("Team").
		Preload("ChecktypeSettings").
		Preload("Programs").First(&policy)
	return &policy, nil
}

func (db vulcanitoStore) DeletePolicy(policy api.Policy) error {
	// TODO: do this on cascade
	result := db.Conn.Delete(&api.ChecktypeSetting{}, "policy_id = ?", policy.ID)
	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}

	result = db.Conn.Delete(&api.ProgramsGroupsPolicies{}, "policy_id = ?", policy.ID)
	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}

	result = db.Conn.Where("team_id = ?", policy.TeamID).Delete(policy)
	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}
	return nil
}

func (db vulcanitoStore) ListChecktypeSetting(policyID string) ([]*api.ChecktypeSetting, error) {
	checktypeSettings := []*api.ChecktypeSetting{}
	res := db.Conn.Find(&checktypeSettings, "policy_id = ?", policyID)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}
	return checktypeSettings, nil
}

func (db vulcanitoStore) CreateChecktypeSetting(setting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	res := db.Conn.Create(&setting)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}
	db.Conn.
		Preload("Policy").
		Preload("Policy.Team").
		Preload("Policy.ChecktypeSettings").
		First(&setting)
	return &setting, nil
}

func (db vulcanitoStore) FindChecktypeSetting(checktypeSettingID string) (*api.ChecktypeSetting, error) {
	checktypeSetting := &api.ChecktypeSetting{ID: checktypeSettingID}
	res := db.Conn.Find(&checktypeSetting)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return checktypeSetting, nil
}

func (db vulcanitoStore) UpdateChecktypeSetting(checktypeSetting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	found := &api.ChecktypeSetting{ID: checktypeSetting.ID}
	res := db.Conn.Find(&found)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	result := db.Conn.Model(&checktypeSetting).Updates(checktypeSetting)
	if result.Error != nil {
		return nil, db.logError(errors.Update(result.Error))
	}
	db.Conn.First(&checktypeSetting)

	return &checktypeSetting, nil
}

func (db vulcanitoStore) DeleteChecktypeSetting(checktypeSettingID string) error {
	checktypeSetting := &api.ChecktypeSetting{ID: checktypeSettingID}

	res := db.Conn.Find(&checktypeSetting)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return db.logError(errors.NotFound(res.Error))
		}
		return db.logError(errors.Database(res.Error))
	}

	res = db.Conn.Delete(checktypeSetting)
	if res.Error != nil {
		return db.logError(errors.Delete(res.Error))
	}

	return nil
}
