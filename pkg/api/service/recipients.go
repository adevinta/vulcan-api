/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) UpdateRecipients(ctx context.Context, teamID string, emails []string) error {
	return s.db.UpdateRecipients(teamID, emails)
}

func (s vulcanitoService) ListRecipients(ctx context.Context, teamID string) ([]*api.Recipient, error) {
	return s.db.ListRecipients(teamID)
}
