/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) StatsMTTR(ctx context.Context, params api.StatsParams) (*api.StatsMTTR, error) {
	return s.vulndbClient.StatsMTTR(ctx, params)
}

func (s vulcanitoService) StatsExposure(ctx context.Context, params api.StatsParams) (*api.StatsExposure, error) {
	return s.vulndbClient.StatsExposure(ctx, params)
}

func (s vulcanitoService) StatsCurrentExposure(ctx context.Context, params api.StatsParams) (*api.StatsCurrentExposure, error) {
	return s.vulndbClient.StatsCurrentExposure(ctx, params)
}

func (s vulcanitoService) StatsOpen(ctx context.Context, params api.StatsParams) (*api.StatsOpen, error) {
	return s.vulndbClient.StatsOpen(ctx, params)
}

func (s vulcanitoService) StatsFixed(ctx context.Context, params api.StatsParams) (*api.StatsFixed, error) {
	return s.vulndbClient.StatsFixed(ctx, params)
}

func (s vulcanitoService) StatsAssets(ctx context.Context, params api.StatsParams) (*api.StatsAssets, error) {
	return s.vulndbClient.StatsAssets(ctx, params)
}

func (s vulcanitoService) StatsCoverage(ctx context.Context, teamID string) (*api.StatsCoverage, error) {
	dg, err := s.db.FindGroupInfo(api.Group{
		Name:   "Default",
		TeamID: teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("Default group not found: %w", err)
	}

	// Find the group id of the Redcon group of the team.
	rg, err := s.db.FindGroupInfo(api.Group{
		Name:   api.DiscoveredAssetsGroupName,
		TeamID: teamID,
	})
	if err != nil {
		if errors.IsRootOfKind(err, errors.ErrNotFound) {
			return &api.StatsCoverage{Coverage: -1}, nil
		}
		return nil, err
	}

	scanned := []string{dg.ID}

	sg, err := s.db.FindGroupInfo(api.Group{
		Name:   "Sensitive",
		TeamID: teamID,
	})
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}

	if err == nil {
		scanned = append(scanned, sg.ID)
	}

	cntScanned, err := s.db.CountAssetsInGroups(teamID, scanned)
	if err != nil {
		return nil, err
	}

	cntAll, err := s.db.CountAssetsInGroups(teamID, append(scanned, rg.ID))
	if err != nil {
		return nil, err
	}
	// Avoid divide 0 when there a no assets in any group. In that case whe
	// consider always to be 100%.
	if cntAll == 0 {
		return &api.StatsCoverage{Coverage: float64(1)}, nil
	}
	return &api.StatsCoverage{Coverage: float64(cntScanned) / float64(cntAll)}, nil
}
