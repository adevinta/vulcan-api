/*
Copyright 2021 Adevinta
*/

package store

import (
	"log"
	"testing"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestVulcanitoStore_CreateProgram(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	program := api.Program{
		ID:     "dd5562a7-d47e-4281-9d2d-e57cbc00b492",
		TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065",
		ProgramsGroupsPolicies: []*api.ProgramsGroupsPolicies{
			&api.ProgramsGroupsPolicies{
				GroupID:  "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
				PolicyID: "0473F67E-E262-4086-BEC5-55CB5071481D",
			},
		},
		Name: "My program",
		Cron: "cron",
	}
	createdProgram, err := testStoreLocal.CreateProgram(program, program.TeamID)
	if err != nil {
		t.Errorf("Cannot create program: %s", err.Error())
		t.FailNow()
	}
	if createdProgram.Name != "My program" {
		t.Error("Wrong name")
	}

	programWrongGroup := api.Program{
		ProgramsGroupsPolicies: []*api.ProgramsGroupsPolicies{
			&api.ProgramsGroupsPolicies{
				GroupID:  "721d1c6b-f559-f559-f559-ca1820173a3c",
				PolicyID: "b69fb157-df28-45b4-8fe5-ea614452e921",
			},
		},
		Name: "My program",
		Cron: "cron",
	}
	programNotCreated, _ := testStoreLocal.CreateProgram(programWrongGroup, "93449fc4-6a84-4058-bac2-200e584ec435")
	if programNotCreated != nil {
		t.Error("Program cannot be created with a group not belonging to the team")
	}
}

func TestVulcanitoStore_UpdateProgram(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	program := api.Program{
		ID:     "1789B7B6-E8ED-49D9-A5A8-9FF9323593B6",
		TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065",
		ProgramsGroupsPolicies: []*api.ProgramsGroupsPolicies{
			&api.ProgramsGroupsPolicies{
				GroupID:  "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
				PolicyID: "0473F67E-E262-4086-BEC5-55CB5071481D",
			},
			&api.ProgramsGroupsPolicies{
				GroupID:  "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
				PolicyID: "48ead64c-fad1-4b18-8d7e-d02bc05dc5f5",
			},
		},
		Name: "updated program",
		Cron: "cron2",
	}
	updatedProgram, err := testStoreLocal.UpdateProgram(program, program.TeamID)
	if err != nil {
		t.Fatalf("Unexpected error updating program %+v", err)
	}
	if updatedProgram.Name != "updated program" {
		t.Fatalf("Unexpected program name after update")
	}
	program.TeamID = "a14c7c65-66ab-4676-bcf6-0dea9719f5c8"
	_, err = testStoreLocal.UpdateProgram(program, program.TeamID)
	if err == nil || !errors.IsRootOfKind(err, errors.ErrNotFound) {
		t.Errorf("Update program should not update programs with wrong teamID")
	}
}

func TestVulcanitoStore_DeleteProgram(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	program := api.Program{
		ID:     "0473F67E-E262-4086-BEC5-55CB5071481D",
		TeamID: "",
	}
	err = testStoreLocal.DeleteProgram(program, "3C7C2963-6A03-4A25-A822-EBEB237DB065")
	if err != nil {
		t.Fatalf("Could not delete program, error %+v", err)
	}

	programOtherTeam := api.Program{
		ID: "849cc1cc-e99c-46f9-83b5-f2241a850280",
	}
	err = testStoreLocal.DeleteProgram(programOtherTeam, "3C7C2963-6A03-4A25-A822-EBEB237DB06")
	if err == nil {
		t.Error("Could delete program not belonging to correct team")
	}
}
