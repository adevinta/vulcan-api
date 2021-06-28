/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/errors"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) ListAssetAnnotations(ctx context.Context, teamID string, assetID string) ([]*api.AssetAnnotation, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}
	if assetID == "" {
		return nil, errors.NotFound(`Asset ID is empty`)
	}

	// Route to store layer
	result, err := s.db.ListAssetAnnotations(teamID, assetID)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

func (s vulcanitoService) CreateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}
	if assetID == "" {
		return nil, errors.NotFound(`Asset ID is empty`)
	}
	if len(annotations) == 0 {
		return nil, errors.NotFound(`Annotations are empty`)
	}

	// Route to store layer
	result, err := s.db.CreateAssetAnnotations(teamID, assetID, annotations)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

func (s vulcanitoService) UpdateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}
	if assetID == "" {
		return nil, errors.NotFound(`Asset ID is empty`)
	}
	if len(annotations) == 0 {
		return nil, errors.NotFound(`Annotations are empty`)
	}

	// Route to store layer
	result, err := s.db.UpdateAssetAnnotations(teamID, assetID, annotations)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

func (s vulcanitoService) PutAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}
	if assetID == "" {
		return nil, errors.NotFound(`Asset ID is empty`)
	}

	// Route to store layer
	result, err := s.db.PutAssetAnnotations(teamID, assetID, annotations)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

func (s vulcanitoService) DeleteAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) error {
	if teamID == "" {
		return errors.NotFound(`Team ID is empty`)
	}
	if assetID == "" {
		return errors.NotFound(`Asset ID is empty`)
	}
	if len(annotations) == 0 {
		return errors.NotFound(`Annotations are empty`)
	}

	// Route to store layer
	err := s.db.DeleteAssetAnnotations(teamID, assetID, annotations)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return err
}
