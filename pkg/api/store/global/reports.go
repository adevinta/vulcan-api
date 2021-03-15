/*
Copyright 2021 Adevinta
*/

package global

func init() {
	registerReport(PeriodicDigestReport)
}

var (
	// PeriodicDigestReport specifies the data for the digest report
	// to be sent on every Wednesday at 8am UTC.
	PeriodicDigestReport = Report{
		ID:              "periodic-digest-report",
		Name:            "Periodic Digest Report",
		DefaultSchedule: "0 8 * * 3",
	}
)
