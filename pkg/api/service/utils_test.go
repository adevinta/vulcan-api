/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	datesFieldNames = []string{"CreatedAt", "UpdatedAt"}
)

// VulcanitoServiceBuilder defines a function that returns a vulcanito service with dependencies properly mocked.
type VulcanitoServiceBuilder func(api.VulcanitoStore) api.VulcanitoService

func buildDefaultVulcanitoSrv(s api.VulcanitoStore, l log.Logger) api.VulcanitoService {
	srv := vulcanitoService{
		db:     s,
		logger: l,
	}
	return srv
}

// VulcanitoServiceTestArgs defines common arguments required for all the vulcanito service tests.
type VulcanitoServiceTestArgs struct {
	BuildSrv VulcanitoServiceBuilder
	Ctx      context.Context
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}
