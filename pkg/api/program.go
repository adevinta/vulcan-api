/*
Copyright 2021 Adevinta
*/

package api

import (
	"errors"
	"fmt"

	vulcanerrors "github.com/adevinta/errors"

	"time"
)

var (
	// ErrInvalidProgramGroupPolicy is returned when any of the groups of
	// policies in a program does not have
	ErrInvalidProgramGroupPolicy = errors.New("the program must have, at least, one asset and one checktype")

	// ErrNoProgramsGroupsPolicies is returned when there are any policy group
	// with, at least, one asset and checktype.
	ErrNoProgramsGroupsPolicies = errors.New("no PoliciesGroups defined in the current program")
)

type Program struct {
	ID                     string `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	TeamID                 string
	Team                   *Team
	ProgramsGroupsPolicies []*ProgramsGroupsPolicies `json:"program_policies" validate:"required"`
	Name                   string                    `json:"name" validate:"required"`
	Cron                   string                    `gorm:"-" json:"cron"` // A program can have empty cron expression, e.g: a program to be run on demand.
	Autosend               *bool                     `json:"autosend"`
	Disabled               *bool                     `json:"disabled"`
	Global                 *bool                     `gorm:"-" json:"global"`
	CreatedAt              *time.Time                `json:"-"`
	UpdatedAt              *time.Time                `json:"-"`
}

// ValidateGroupsPolicies validates that at least one of the groups policies in
// a program have, at least, one asset and one checktype.
func (p Program) ValidateGroupsPolicies() error {
	if len(p.ProgramsGroupsPolicies) < 1 {
		return vulcanerrors.Validation(fmt.Errorf("%w: %v ", ErrNoProgramsGroupsPolicies, p.ID))
	}
	var err error
	for _, gp := range p.ProgramsGroupsPolicies {
		if err = gp.Validate(); err == nil {
			return nil
		}
	}
	return err
}

// ProgramsGroupsPolicies defines the association between a group and a policy in a
// program.
type ProgramsGroupsPolicies struct {
	ProgramID string `gorm:"primary_key" json:"program_id" validate:"required"`
	Program   *Program
	PolicyID  string `gorm:"primary_key" json:"policy_id" validate:"required"`
	Policy    *Policy
	GroupID   string `gorm:"primary_key" json:"group_id" validate:"required"`
	Group     *Group
}

// Validate that the ProgramsGroupsPolicies have, at least, one asset and one
// checktype in the groups policies list.
func (p ProgramsGroupsPolicies) Validate() error {

	if p.Group == nil || len(p.Group.AssetGroup) < 1 {
		return vulcanerrors.Validation(ErrInvalidProgramGroupPolicy)
	}

	if p.Policy == nil || len(p.Policy.ChecktypeSettings) < 1 {
		return vulcanerrors.Validation(ErrInvalidProgramGroupPolicy)
	}
	return nil
}

type ProgramResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Global       bool             `json:"global"`
	Schedule     ScheduleResponse `json:"schedule"`
	Autosend     bool             `json:"autosend"`
	Disabled     bool             `json:"disabled"`
	PolicyGroups []PolicyGroup    `json:"policy_groups"`
}

type PolicyGroup struct {
	Group  *GroupResponse  `json:"group"`
	Policy *PolicyResponse `json:"policy"`
}

type ScheduleResponse struct {
	Cron string `json:"cron"`
}

func (p Program) ToResponse() *ProgramResponse {
	var autosend bool
	// Autosend should be never nil this is just to avoid panics if for whatever
	// reason this precondition is not meet.
	if p.Autosend != nil {
		autosend = *p.Autosend
	}

	var disabled bool
	// Disabled should be never nil this is just to avoid panics if for whatever
	// reason this precondition is not meet.
	if p.Disabled != nil {
		disabled = *p.Disabled
	}

	var global bool
	if p.Global != nil {
		global = *p.Global
	}
	response := ProgramResponse{
		ID:           p.ID,
		Name:         p.Name,
		Schedule:     ScheduleResponse{Cron: p.Cron},
		Autosend:     autosend,
		Disabled:     disabled,
		PolicyGroups: []PolicyGroup{},
		Global:       global,
	}
	for _, v := range p.ProgramsGroupsPolicies {
		v := v
		if v == nil {
			continue
		}
		policyGroup := PolicyGroup{}
		if v.Group != nil {
			policyGroup.Group = v.Group.ToResponse()
		}
		if v.Policy != nil {
			policyGroup.Policy = v.Policy.ToResponse()
		}
		response.PolicyGroups = append(response.PolicyGroups, policyGroup)
	}

	return &response
}

// GlobalProgramsMetadata defines the shape of the metadata stored
// per team for a given global program.
type GlobalProgramsMetadata struct {
	TeamID    string `gorm:"primary_key"`
	Program   string `gorm:"primary_key"`
	Autosend  *bool
	Disabled  *bool
	Cron      string `gorm:"-" json:"cron"` // A program can have empty cron expression, e.g: a program to be run on demand.
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
