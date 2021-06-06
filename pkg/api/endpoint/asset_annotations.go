/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
)

type AssetAnnotationRequest struct {
	TeamID      string            `json:"team_id" urlvar:"team_id"`
	AssetID     string            `json:"asset_id" urlvar:"asset_id"`
	Annotations map[string]string `json:"annotations"`
}

func makeListAssetAnnotationsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Handle input
		req, ok := request.(*AssetAnnotationRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Route to service layer
		annotations, err := s.ListAssetAnnotations(ctx, req.TeamID, req.AssetID)
		if err != nil {
			return nil, err
		}

		// Merge annotations into one map
		response := api.AssetAnnotationResponse{}
		for _, annotation := range annotations {
			response[annotation.Key] = annotation.Value
		}

		return Ok{response}, nil
	}
}

func makeCreateAssetAnnotationsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Handle input
		req, ok := request.(*AssetAnnotationRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Transform input from standardized "map[string]string" format to the
		// internal AssetAnnotation model
		annotations := []*api.AssetAnnotation{}
		for k, v := range req.Annotations {
			annotations = append(annotations, &api.AssetAnnotation{
				Key:   k,
				Value: v,
			})
		}

		// Route to service layer
		newAnnotations, err := s.CreateAssetAnnotations(ctx, req.TeamID, req.AssetID, annotations)
		if err != nil {
			return nil, err
		}

		// Format response
		response := api.AssetAnnotationResponse{}
		for _, annotation := range newAnnotations {
			response[annotation.Key] = annotation.Value
		}

		return Ok{response}, nil
	}
}

func makeUpdateAssetAnnotationsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Handle input
		req, ok := request.(*AssetAnnotationRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Transform input from standardized "map[string]string" format to the
		// internal AssetAnnotation model
		annotations := []*api.AssetAnnotation{}
		for k, v := range req.Annotations {
			annotations = append(annotations, &api.AssetAnnotation{
				Key:   k,
				Value: v,
			})
		}

		// Route to service layer
		newAnnotations, err := s.UpdateAssetAnnotations(ctx, req.TeamID, req.AssetID, annotations)
		if err != nil {
			return nil, err
		}

		// Merge annotations into one map
		response := api.AssetAnnotationResponse{}
		for _, annotation := range newAnnotations {
			response[annotation.Key] = annotation.Value
		}

		return Ok{response}, nil
	}
}

type AssetAnnotationDeleteRequest struct {
	TeamID      string   `json:"team_id" urlvar:"team_id"`
	AssetID     string   `json:"asset_id" urlvar:"asset_id"`
	Annotations []string `json:"annotations"`
}

func makeDeleteAssetAnnotationsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Handle input
		req, ok := request.(*AssetAnnotationDeleteRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Transform input from standardized map[string]string to AssetAnnotation
		// model
		annotations := []*api.AssetAnnotation{}
		for _, k := range req.Annotations {
			annotations = append(annotations, &api.AssetAnnotation{Key: k})
		}

		// Route to service layer
		newAnnotations, err := s.DeleteAssetAnnotations(ctx, req.TeamID, req.AssetID, annotations)
		if err != nil {
			return nil, err
		}

		// Merge annotations into one map
		response := api.AssetAnnotationResponse{}
		for _, annotation := range newAnnotations {
			response[annotation.Key] = annotation.Value
		}

		return Ok{response}, nil
	}
}
