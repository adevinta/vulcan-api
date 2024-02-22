/*
Copyright 2021 Adevinta
*/

package service

import (
	"github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/reports"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/tickets"
	"github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"
	metrics "github.com/adevinta/vulcan-metrics-client"
)

// AWSAccounts defines the services realted to AWS Accounts required by the
// Vulcan API.
type AWSAccounts interface {
	Name(AccountID string) (string, error)
}

// vulcanitoService implements VulcanitoService
type vulcanitoService struct {
	jwtConfig             jwt.Config
	db                    api.VulcanitoStore
	logger                log.Logger
	programScheduler      schedule.ScanScheduler
	scanEngineConfig      scanengine.Config
	reportsConfig         reports.Config
	vulndbClient          vulnerabilitydb.Client
	vulcantrackerClient   tickets.Client
	reportsClient         *reports.Client
	metricsClient         metrics.Client
	awsAccounts           AWSAccounts
	allowedTrackerTeams   []string // feature flag.
	DNSHostnameValidation bool
}

//go:generate impl -output logging.go -stub templates/logging/impl.tmpl -header templates/logging/header.tmpl "middleware loggingMiddleware" api.VulcanitoService
//go:generate gofmt -w logging.go

// New returns a basic Service with all of the expected middlewares wired in.
func New(logger log.Logger, db api.VulcanitoStore, jwtConfig jwt.Config,
	scanEngineConfig scanengine.Config, programScheduler schedule.ScanScheduler, reportsConfig reports.Config,
	vulndbClient vulnerabilitydb.Client, vulcantrackerClient tickets.Client, reportsClient *reports.Client,
	metricsClient metrics.Client, awsAccounts AWSAccounts, allowedTrackerTeams []string, DNSHostnameValidation bool) api.VulcanitoService {

	var svc api.VulcanitoService
	{
		svc = vulcanitoService{db: db,
			jwtConfig:             jwtConfig,
			logger:                logger,
			scanEngineConfig:      scanEngineConfig,
			programScheduler:      programScheduler,
			reportsConfig:         reportsConfig,
			vulndbClient:          vulndbClient,
			vulcantrackerClient:   vulcantrackerClient,
			reportsClient:         reportsClient,
			metricsClient:         metricsClient,
			awsAccounts:           awsAccounts,
			allowedTrackerTeams:   allowedTrackerTeams,
			DNSHostnameValidation: DNSHostnameValidation,
		}
	}
	return LoggingMiddleware(logger)(svc)
}
