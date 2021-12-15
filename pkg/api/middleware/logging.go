/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/transport/http"
	kithttp "github.com/go-kit/kit/transport/http"
	uuid "github.com/satori/go.uuid"

	"github.com/adevinta/vulcan-api/pkg/api"
	vulcanendpoint "github.com/adevinta/vulcan-api/pkg/api/endpoint"
)

func EndpointLogging(logger log.Logger, name string, db api.VulcanitoStore) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			// Avoid logging on healthcheck
			if name == vulcanendpoint.Healthcheck {
				return next(ctx, request)
			}

			var XRequestID, u, team, teamName string

			if ctx != nil {
				XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
				user, ok := ctx.Value("email").(api.User)
				if ok {
					u = user.Email
				}

				path, ok := ctx.Value(http.ContextKeyRequestPath).(string)
				if ok {
					teamsPrefix := "/api/v1/teams/"
					if strings.HasPrefix(path, teamsPrefix) {
						path = path[len(teamsPrefix):]
						if len(path) >= 36 {
							_, err := uuid.FromString(path[:36])
							if err == nil {
								team = path[:36]

								t, dbErr := db.FindTeam(team)
								if dbErr != nil {
									_ = level.Debug(logger).Log(
										"X-Request-ID", XRequestID,
										"endpoint", name,
										"msg", "error retrieving team name",
										"request", fmt.Sprintf("%+v", request),
										"user", u,
										"team", team,
										"team-name", teamName,
										"error", fmt.Sprintf("%+v", dbErr),
									)
								} else {
									teamName = t.Name
								}
							}
						}
					}
				}
			}

			begin := time.Now()
			_ = level.Debug(logger).Log(
				"X-Request-ID", XRequestID,
				"endpoint", name,
				"msg", "calling endpoint",
				"request", fmt.Sprintf("%+v", request),
				"user", u,
				"team", team,
				"team-name", teamName,
			)
			response, err := next(ctx, request)
			_ = level.Debug(logger).Log(
				"X-Request-ID", XRequestID,
				"endpoint", name,
				"msg", "called endpoint",
				"response", fmt.Sprintf("%+v", response),
				"took", time.Since(begin),
				"err", err,
				"user", u,
				"team", team,
				"team-name", teamName,
			)

			return response, err
		}
	}
}
