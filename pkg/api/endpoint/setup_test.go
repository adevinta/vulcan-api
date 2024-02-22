/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"os"

	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/adevinta/vulcan-api/pkg/reports"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/lib/pq"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/service"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	ignoreFieldsAssetResponse = cmpopts.IgnoreFields(api.AssetResponse{}, "AssetType")
)

var svcLogger kitlog.Logger

func init() {
	svcLogger = kitlog.NewLogfmtLogger(os.Stderr)
	svcLogger = kitlog.With(svcLogger, "ts", kitlog.DefaultTimestampUTC)
	svcLogger = kitlog.With(svcLogger, "caller", kitlog.DefaultCaller)
}

type localScheduler struct {
	schedules map[string]string
}

func (l *localScheduler) CreateScanSchedule(programID, teamID, cronExpr string) error {
	l.schedules[programID] = cronExpr
	return nil
}

func (l *localScheduler) BulkCreateScanSchedules(schedules []schedule.ScanBulkSchedule) error {
	return nil
}

func (l *localScheduler) GetScanScheduleByID(programID string) (string, error) {
	return l.schedules[programID], nil
}

func (l *localScheduler) DeleteScanSchedule(programID string) error {
	delete(l.schedules, programID)
	return nil
}

func buildTestService(testStore api.VulcanitoStore) api.VulcanitoService {
	s := &localScheduler{
		schedules: make(map[string]string),
	}
	return service.New(svcLogger, testStore, jwt.Config{}, scanengine.Config{Url: ""},
		s, reports.Config{}, vulnerabilitydb.NewClient(nil, "", true),
		nil, nil, nil, awscatalogue.NewAWSAccounts(nil, nil), []string{},
		false)
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}
