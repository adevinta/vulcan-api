/*
Copyright 2024 Adevinta
*/

package service

import (
	"context"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/tickets"
)

func TestVulcanitoService_OnboardedTeam(t *testing.T) {
	srv := vulcanitoService{
		vulcantrackerClient: tickets.NewClient(nil, "", true),
	}
	ctx := context.Background()
	ctxAdmin := api.ContextWithUser(ctx, api.User{Admin: common.Bool(true)})
	ctxObserver := api.ContextWithUser(ctx, api.User{Observer: common.Bool(true)})

	tests := []struct {
		name           string
		srv            vulcanitoService
		context        context.Context
		onboardedTeams []string
		teamID         string
		want           bool
	}{
		{
			name:           "OnboardedAllTeamsNoClient",
			srv:            vulcanitoService{},
			context:        ctxAdmin,
			onboardedTeams: []string{"whatever-team", "*"},
			teamID:         "other-team",
			want:           false,
		},
		{
			name:           "HappyPath-admin",
			srv:            srv,
			context:        ctxAdmin,
			onboardedTeams: []string{"whatever-team"},
			teamID:         "whatever-team",
			want:           true,
		},
		{
			name:           "HappyPath-observer",
			srv:            srv,
			context:        ctxObserver,
			onboardedTeams: []string{"whatever-team"},
			teamID:         "whatever-team",
			want:           true,
		},
		{
			name:           "HappyPath-no-auth",
			srv:            srv,
			context:        ctx,
			onboardedTeams: []string{"whatever-team"},
			teamID:         "whatever-team",
			want:           false,
		},
		{
			name:           "NoOnboardedTeam",
			srv:            srv,
			context:        ctxAdmin,
			onboardedTeams: []string{"whatever-team"},
			teamID:         "other-team",
			want:           false,
		},
		{
			name:           "OnboardedAllTeams",
			srv:            srv,
			context:        ctxAdmin,
			onboardedTeams: []string{"whatever-team", "*"},
			teamID:         "other-team",
			want:           true,
		},
		{
			name:           "NoOnboardedAllTeams",
			srv:            srv,
			context:        ctxAdmin,
			onboardedTeams: []string{},
			teamID:         "other-team",
			want:           false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.srv.IsATeamOnboardedInVulcanTracker(tt.context, tt.teamID, tt.onboardedTeams) != tt.want {
				t.Fatal("error")
			}
		})
	}
}
