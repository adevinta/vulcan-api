/*
Copyright 2022 Adevinta
*/
package endpoint

import (
	"context"
	"errors"

	aerrors "github.com/adevinta/errors"

	"github.com/adevinta/vulcan-api/pkg/securitygraph"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
)

// IntelAPIClient defines the interface of the Intel API client
// needed by the Vulcan API to expose them acting as a Gateway.
type IntelAPIClient interface {
	BlastRadius(req securitygraph.BlastRadiusRequest) (securitygraph.BlastRadiusResponse, error)
}

func makeBlastRadiusEndpoint(i IntelAPIClient, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*securitygraph.BlastRadiusRequest)
		if !ok {
			return nil, aerrors.Assertion("Type assertion failed")
		}
		response, err := i.BlastRadius(*req)
		if err == nil {
			return Ok{response}, nil
		}
		statusErr := securitygraph.HttpStatusError{}
		if errors.As(err, &statusErr) {
			resError := aerrors.Error{
				HTTPStatusCode: statusErr.Status,
				Message:        statusErr.Msg,
			}
			return nil, resError
		}
		return nil, err
	}
}
