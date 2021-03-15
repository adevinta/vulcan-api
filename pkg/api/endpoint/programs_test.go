/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestMakeCreateProgramsEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	// var False = false
	var True = true

	tests := []struct {
		req            interface{}
		name           string
		want           interface{}
		wantErr        error
		wantClassified bool
	}{
		{
			name: "HappyPath",
			req: &ProgramRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Name:   "New Program",
				PolicyGroups: &ProgramPolicyGroups{
					ProgramsPolicyGroup{
						PolicyID: "0473F67E-E262-4086-BEC5-55CB5071481D",
						GroupID:  "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					},
				},
			},
			want: Created{
				&api.ProgramResponse{
					ID:       "eb8556b0-14ee-449f-b60c-e6876eec07d5",
					Name:     "New Program",
					Global:   false,
					Disabled: false,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ProgramWithDisabledTrue",
			req: &ProgramRequest{
				TeamID:   "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Name:     "New Program",
				Disabled: &True,
				PolicyGroups: &ProgramPolicyGroups{
					ProgramsPolicyGroup{
						PolicyID: "0473F67E-E262-4086-BEC5-55CB5071481D",
						GroupID:  "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					},
				},
			},
			want: Created{
				&api.ProgramResponse{
					ID:       "eb8556b0-14ee-449f-b60c-e6876eec07d5",
					Name:     "New Program",
					Global:   false,
					Disabled: true,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeCreateProgramEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.ProgramResponse{}, "ID")})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestMakeUpdateProgramsEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	var False = false
	var True = true

	tests := []struct {
		req            interface{}
		name           string
		want           interface{}
		wantErr        error
		wantClassified bool
	}{
		{
			name: "KeepDisabledFieldOriginalValue-1",
			req: &ProgramRequest{
				TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065",
				ID:     "8d30cb6d-0cf6-4c54-8ec1-1e7c4249e094",
			},
			want: Ok{
				&api.ProgramResponse{
					ID:       "eb8556b0-14ee-449f-b60c-e6876eec07d5",
					Name:     "Enabled Program",
					Global:   false,
					Schedule: api.ScheduleResponse{Cron: "time"},
					Disabled: false,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "KeepDisabledFieldOriginalValue-2",
			req: &ProgramRequest{
				TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065",
				ID:     "abd059d0-f81a-433b-be8a-ace3ebb6c926",
			},
			want: Ok{
				&api.ProgramResponse{
					ID:       "abd059d0-f81a-433b-be8a-ace3ebb6c926",
					Name:     "Disabled Program",
					Global:   false,
					Schedule: api.ScheduleResponse{Cron: "time"},
					Disabled: true,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "DisableEnabledProgram",
			req: &ProgramRequest{
				TeamID:   "3C7C2963-6A03-4A25-A822-EBEB237DB065",
				ID:       "8d30cb6d-0cf6-4c54-8ec1-1e7c4249e094",
				Disabled: &True,
			},
			want: Ok{
				&api.ProgramResponse{
					ID:       "eb8556b0-14ee-449f-b60c-e6876eec07d5",
					Name:     "Enabled Program",
					Global:   false,
					Schedule: api.ScheduleResponse{Cron: "time"},
					Disabled: true,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "EnableDisabledProgram",
			req: &ProgramRequest{
				TeamID:   "3C7C2963-6A03-4A25-A822-EBEB237DB065",
				ID:       "abd059d0-f81a-433b-be8a-ace3ebb6c926",
				Disabled: &False,
			},
			want: Ok{
				&api.ProgramResponse{
					ID:       "abd059d0-f81a-433b-be8a-ace3ebb6c926",
					Name:     "Disabled Program",
					Global:   false,
					Schedule: api.ScheduleResponse{Cron: "time"},
					Disabled: false,
					PolicyGroups: []api.PolicyGroup{
						api.PolicyGroup{
							Group:  &api.GroupResponse{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "mygroup", AssetsCount: intPointer(1)},
							Policy: &api.PolicyResponse{ID: "0473f67e-e262-4086-bec5-55cb5071481d", Name: "my policy"},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeUpdateProgramEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.ProgramResponse{}, "ID")})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func intPointer(i int) *int {
	return &i
}
