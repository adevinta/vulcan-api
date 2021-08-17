/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adevinta/errors"
	metrics "github.com/adevinta/vulcan-metrics-client"
	stdjwt "github.com/dgrijalva/jwt-go"
	jwtkit "github.com/go-kit/kit/auth/jwt"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/adevinta/vulcan-api/pkg/api/endpoint"
	"github.com/adevinta/vulcan-api/pkg/api/transport"
)

const (
	// Metric names
	metricTotal    = "vulcan.request.total"
	metricDuration = "vulcan.request.duration"
	metricFailed   = "vulcan.request.failed"

	// Tags
	tagComponent = "component"
	tagAction    = "action"
	tagEntity    = "entity"
	tagMethod    = "method"
	tagStatus    = "status"
	tagUser      = "user"
	tagTeam      = "team"

	// Entities
	entityUser      = "user"
	entityTeam      = "team"
	entityRecipient = "recipient"
	entityAsset     = "asset"
	entityProgram   = "program"
	entityPolicy    = "policiy"
	entityCheck     = "check"
	entityScan      = "scan"
	entityReport    = "report"
	entityFinding   = "finding"
	entityStats     = "stats"

	apiComponent  = "api"
	unknownAction = "unknown"
)

var (
	endpointToEntity = map[string]string{
		// User
		endpoint.ListUsers:        entityUser,
		endpoint.CreateUser:       entityUser,
		endpoint.UpdateUser:       entityUser,
		endpoint.FindUser:         entityUser,
		endpoint.DeleteUser:       entityUser,
		endpoint.FindProfile:      entityUser,
		endpoint.GenerateAPIToken: entityUser,
		endpoint.ListTeamMembers:  entityUser,
		endpoint.FindTeamMember:   entityUser,
		endpoint.CreateTeamMember: entityUser,
		endpoint.UpdateTeamMember: entityUser,
		endpoint.DeleteTeamMember: entityUser,
		// Team
		endpoint.CreateTeam:      entityTeam,
		endpoint.UpdateTeam:      entityTeam,
		endpoint.FindTeam:        entityTeam,
		endpoint.ListTeams:       entityTeam,
		endpoint.DeleteTeam:      entityTeam,
		endpoint.FindTeamsByUser: entityTeam,
		// Recipient
		endpoint.ListRecipients:   entityRecipient,
		endpoint.UpdateRecipients: entityRecipient,
		// Asset
		endpoint.ListAssets:             entityAsset,
		endpoint.CreateAsset:            entityAsset,
		endpoint.CreateAssetMultiStatus: entityAsset,
		endpoint.FindAsset:              entityAsset,
		endpoint.UpdateAsset:            entityAsset,
		endpoint.DeleteAsset:            entityAsset,
		endpoint.CreateGroup:            entityAsset,
		endpoint.ListGroups:             entityAsset,
		endpoint.UpdateGroup:            entityAsset,
		endpoint.DeleteGroup:            entityAsset,
		endpoint.FindGroup:              entityAsset,
		endpoint.GroupAsset:             entityAsset,
		endpoint.UngroupAsset:           entityAsset,
		endpoint.ListAssetGroup:         entityAsset,
		// Program
		endpoint.ListPrograms:          entityProgram,
		endpoint.CreateProgram:         entityProgram,
		endpoint.FindProgram:           entityProgram,
		endpoint.UpdateProgram:         entityProgram,
		endpoint.DeleteProgram:         entityProgram,
		endpoint.CreateSchedule:        entityProgram,
		endpoint.DeleteSchedule:        entityProgram,
		endpoint.ScheduleGlobalProgram: entityProgram,
		// Policy
		endpoint.ListPolicies: entityPolicy,
		endpoint.CreatePolicy: entityPolicy,
		endpoint.FindPolicy:   entityPolicy,
		endpoint.UpdatePolicy: entityPolicy,
		endpoint.DeletePolicy: entityPolicy,
		// Check
		endpoint.ListChecktypeSetting:   entityCheck,
		endpoint.CreateChecktypeSetting: entityCheck,
		endpoint.FindChecktypeSetting:   entityCheck,
		endpoint.UpdateChecktypeSetting: entityCheck,
		endpoint.DeleteChecktypeSetting: entityCheck,
		// Scan
		endpoint.ListProgramScans: entityScan,
		endpoint.CreateScan:       entityScan,
		endpoint.FindScan:         entityScan,
		endpoint.AbortScan:        entityScan,
		// Report
		endpoint.FindReport:      entityReport,
		endpoint.CreateReport:    entityReport,
		endpoint.SendReport:      entityReport,
		endpoint.FindReportEmail: entityReport,
		// Finding
		endpoint.ListFindings:           entityFinding,
		endpoint.ListFindingsIssues:     entityFinding,
		endpoint.ListFindingsTargets:    entityFinding,
		endpoint.ListFindingsByIssue:    entityFinding,
		endpoint.ListFindingsByTarget:   entityFinding,
		endpoint.FindFinding:            entityFinding,
		endpoint.CreateFindingOverwrite: entityFinding,
		endpoint.ListFindingOverwrites:  entityFinding,
		endpoint.ListFindingsLabels:     entityFinding,
		// Stats
		endpoint.StatsCoverage:   entityStats,
		endpoint.StatsMTTR:       entityStats,
		endpoint.StatsOpen:       entityStats,
		endpoint.StatsFixed:      entityStats,
		endpoint.GlobalStatsMTTR: entityStats,
	}
)

// MetricsMiddleware implements a metrics middleware over an endpoint.
type MetricsMiddleware interface {
	Measure(next kitendpoint.Endpoint) kitendpoint.Endpoint
}

type metricsMiddleware struct {
	metricsClient metrics.Client
}

// NewMetricsMiddleware creates a new metrics middleware pushing the
// metrics through the given metrics client.
func NewMetricsMiddleware(metricsClient metrics.Client) MetricsMiddleware {
	return &metricsMiddleware{
		metricsClient: metricsClient,
	}
}

func (m *metricsMiddleware) Measure(next kitendpoint.Endpoint) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Time and execute request
		reqStart := time.Now()
		res, err := next(ctx, request)
		reqEnd := time.Now()

		// Collect metrics
		endpoint := ctx.Value(transport.ContextKeyEndpoint).(string)
		httpMethod := ctx.Value(kithttp.ContextKeyRequestMethod).(string)
		httpStatus := parseHTTPStatus(res, err)
		duration := reqEnd.Sub(reqStart).Milliseconds()
		failed := httpStatus >= 400
		user := parseUser(ctx)
		team := parseTeam(ctx)

		// Build tags
		tags := []string{
			fmt.Sprint(tagComponent, ":", apiComponent),
			fmt.Sprint(tagAction, ":", endpoint),
			fmt.Sprint(tagEntity, ":", endpointToEntity[endpoint]),
			fmt.Sprint(tagMethod, ":", httpMethod),
			fmt.Sprint(tagStatus, ":", httpStatus),
			fmt.Sprint(tagUser, ":", user),
			fmt.Sprint(tagTeam, ":", team),
		}

		// Push metrics
		m.pushMetrics(httpMethod, duration, failed, tags)

		return res, err
	}
}

func (m *metricsMiddleware) pushMetrics(httpMethod string, duration int64, failed bool, tags []string) {
	mm := []metrics.Metric{
		{
			Name:  metricTotal,
			Typ:   metrics.Count,
			Value: 1,
			Tags:  tags,
		},
		{
			Name:  metricDuration,
			Typ:   metrics.Histogram,
			Value: float64(duration),
			Tags:  tags,
		},
	}
	if failed {
		mm = append(mm, metrics.Metric{
			Name:  metricFailed,
			Typ:   metrics.Count,
			Value: 1,
			Tags:  tags,
		})
	}

	for _, met := range mm {
		m.metricsClient.Push(met)
	}
}

func parseHTTPStatus(resp interface{}, err error) int {
	// If err is not nil, try to cast to ErrStack and
	// return its StatusCode.
	// Otherwise default to HTTP 500 status code.
	if err != nil {
		if errStack, ok := err.(*errors.ErrorStack); ok {
			return errStack.StatusCode()
		}
		return http.StatusInternalServerError
	}

	// If err is nil, try to cast to endpoint HTTPResponse
	// and return its StatusCode.
	// Otherwise default to HTTP 200 status code.
	if httpResp, ok := resp.(endpoint.HTTPResponse); ok {
		return httpResp.StatusCode()
	}
	return http.StatusOK
}

// parseUser returns the user from the JWT token.
func parseUser(ctx context.Context) string {
	user := ""
	token, ok := ctx.Value(jwtkit.JWTTokenContextKey).(string)
	if ok {
		claims := stdjwt.MapClaims{}

		// IMPORTANT: This trusts the user provided in the JWT token without
		// verifying its signature. This is only used for metrics purposes.
		_, _, err := new(stdjwt.Parser).ParseUnverified(token, claims)
		if err == nil {
			user = claims["sub"].(string)
		}
	}
	return user
}

// parseTeam team returns the team from the URL path.
func parseTeam(ctx context.Context) string {
	team := ""
	httpPath := ctx.Value(kithttp.ContextKeyRequestPath).(string)
	if strings.HasPrefix(httpPath, "/api/v1/teams/") {
		team = strings.Split(httpPath, "/")[4]
	}
	return team
}
