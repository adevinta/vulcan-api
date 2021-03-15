/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/jinzhu/gorm"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) ListPrograms(teamID string) ([]*api.Program, error) {
	programs := []*api.Program{}
	result := db.Conn.Where("team_id = ? ", teamID).
		Preload("ProgramsGroupsPolicies").
		Preload("ProgramsGroupsPolicies.Group").
		Preload("ProgramsGroupsPolicies.Group.Team").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup").
		Preload("ProgramsGroupsPolicies.Policy").
		Preload("ProgramsGroupsPolicies.Policy.Team").
		Preload("ProgramsGroupsPolicies.Policy.ChecktypeSettings").
		Find(&programs)

	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return programs, nil
}

func (db vulcanitoStore) CreateProgram(program api.Program, teamID string) (*api.Program, error) {
	// Ensure default value
	var False = false
	if program.Disabled == nil {
		program.Disabled = &False
	}

	// We have foreign keys defined in the program_policies_group
	// relations no need to check if the policy and the group exist.

	res := db.Conn.Create(&program)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}

	db.Conn.
		Preload("ProgramsGroupsPolicies").
		Preload("ProgramsGroupsPolicies.Group").
		Preload("ProgramsGroupsPolicies.Group.Team").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset.AssetType").
		Preload("ProgramsGroupsPolicies.Policy").
		Preload("ProgramsGroupsPolicies.Policy.Team").
		Preload("ProgramsGroupsPolicies.Policy.ChecktypeSettings").
		First(&program)

	return &program, nil
}

func (db vulcanitoStore) FindProgram(programID string, teamID string) (*api.Program, error) {
	program := &api.Program{}
	result := db.Conn.
		Preload("ProgramsGroupsPolicies").
		Preload("ProgramsGroupsPolicies.Group").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset.AssetType").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset.Team").
		Preload("ProgramsGroupsPolicies.Policy").
		Preload("ProgramsGroupsPolicies.Policy.Team").
		Preload("ProgramsGroupsPolicies.Policy.ChecktypeSettings").
		Find(&program, "id = ? and team_id = ?", programID, teamID)

	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return program, nil
}

func (db vulcanitoStore) removeGroupsPoliciesForProgram(tx *gorm.DB, programID string) error {
	result := tx.Exec("DELETE from programs_groups_policies where program_id = ?", programID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Update(result.Error))
	}
	return nil
}

func (db vulcanitoStore) UpdateProgram(program api.Program, teamID string) (*api.Program, error) {
	if len(program.ProgramsGroupsPolicies) > 0 {
		ok, err := db.checkProgramPolicyGroupsTeam(teamID, program.ProgramsGroupsPolicies)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, db.logError(errors.NotFound(""))
		}
	}
	tx := db.Conn.Begin()
	// If the user specifies a list of GroupsPolicies that list
	// must replace the current one so we will delete the current ones
	// before executing the update.
	if len(program.ProgramsGroupsPolicies) > 1 {
		if err := db.removeGroupsPoliciesForProgram(tx, program.ID); err != nil {
			return nil, err
		}
	}
	result := tx.Model(&program).Where("team_id = ?", teamID).Update(program)
	if result.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Update(result.Error))
	}

	err := tx.
		Preload("ProgramsGroupsPolicies").
		Preload("ProgramsGroupsPolicies.Group").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset.AssetType").
		Preload("ProgramsGroupsPolicies.Group.AssetGroup.Asset.Team").
		Preload("ProgramsGroupsPolicies.Policy").
		Preload("ProgramsGroupsPolicies.Policy.Team").
		Preload("ProgramsGroupsPolicies.Policy.ChecktypeSettings").First(&program).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.Database(err)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Database(err)
	}
	return &program, nil
}

func (db vulcanitoStore) checkProgramPolicyGroupsTeam(teamID string, PGroups []*api.ProgramsGroupsPolicies) (bool, error) {
	if len(PGroups) < 1 {
		return true, nil
	}
	pIDs := []string{}
	gIDs := []string{}
	distinctGroups := map[string]struct{}{}
	distinctPolicies := map[string]struct{}{}
	for _, p := range PGroups {
		p := p
		_, ok := distinctGroups[p.GroupID]
		if !ok {
			distinctGroups[p.GroupID] = struct{}{}
			gIDs = append(gIDs, p.GroupID)
		}
		_, ok = distinctPolicies[p.PolicyID]
		if !ok {
			distinctPolicies[p.PolicyID] = struct{}{}
			pIDs = append(pIDs, p.PolicyID)
		}
	}
	policiesClause := `SELECT count(*) as count FROM policies p JOIN teams t ON
	p.team_id = t.id AND t.id = ? AND p.id IN (?)`
	rows, err := db.Conn.Raw(policiesClause, teamID, pIDs).Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var ret int
	if rows.Next() {
		if err := rows.Scan(&ret); err != nil {
			return false, err
		}
	}
	count := len(pIDs)
	if ret != count {
		return false, nil
	}

	groupsClause := `SELECT count(*) as count FROM groups g JOIN teams t ON
	g.team_id = t.id AND t.id = ? AND g.id IN (?)`
	gRows, err := db.Conn.Raw(groupsClause, teamID, gIDs).Rows()
	if err != nil {
		return false, err
	}
	defer gRows.Close()
	ret = 0
	if gRows.Next() {
		if err := gRows.Scan(&ret); err != nil {
			return false, err
		}
	}
	count = len(gIDs)
	if ret != count {
		return false, nil
	}
	return true, nil
}
func (db vulcanitoStore) DeleteProgram(program api.Program, teamID string) error {
	result := db.Conn.Where("team_id = ?", teamID).Delete(program)
	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}
	return nil
}
