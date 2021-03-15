/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
)

type HealthcheckJSONRequest struct {
}

func makeHealthcheckEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var hc api.Healthcheck

		err = svc.Healthcheck(ctx)
		if err != nil {
			return nil, err
		}

		hc.Status = "OK"
		return Ok{hc.ToResponse()}, nil
	}
}
