/*
Copyright 2021 Adevinta
*/

package schedule

type ScanScheduler interface {
	CreateScanSchedule(programID, teamID, cronExpr string) error
	GetScanScheduleByID(programID string) (string, error)
	DeleteScanSchedule(programID string) error
	BulkCreateScanSchedules(schedules []ScanBulkSchedule) error
}

type ReportScheduler interface {
	CreateReportSchedule(teamID, cronExpr string) error
	GetReportScheduleByID(teamID string) (string, error)
	DeleteReportSchedule(teamID string) error
	BulkCreateReportSchedules(schedules []ReportBulkSchedule) error
}
