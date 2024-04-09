/*
Copyright 2024 Adevinta
*/

package service

import (
	"context"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/tickets"
)

func TestVulcanitoService_OnboardedTeam(t *testing.T) {
	srv := vulcanitoService{}
	if srv.IsATeamOnboardedInVulcanTracker(context.TODO(), "whatever-team", []string{"whatever-team"}) {
		t.Fatal("vulcanitoService.IsATeamOnboardedInVulcanTracker() no-client")
	}
	srv.vulcantrackerClient = tickets.NewClient(nil, "", true)
	if !srv.IsATeamOnboardedInVulcanTracker(context.TODO(), "whatever-team", []string{"*"}) {
		t.Fatal("vulcanitoService.IsATeamOnboardedInVulcanTracker() all")
	}
	if !srv.IsATeamOnboardedInVulcanTracker(context.TODO(), "whatever-team", []string{"whatever-team"}) {
		t.Fatal("vulcanitoService.IsATeamOnboardedInVulcanTracker() explicit")
	}
	if srv.IsATeamOnboardedInVulcanTracker(context.TODO(), "whatever-team", []string{"other"}) {
		t.Fatal("vulcanitoService.IsATeamOnboardedInVulcanTracker() not")
	}
}
