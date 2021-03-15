/*
Copyright 2021 Adevinta
*/

package global

import "github.com/adevinta/vulcan-api/pkg/api"

func init() {
	registerProgram(PeriodicFullScan)
	registerProgram(RedconScan)
	registerProgram(WebScanning)
}

var vFalse = false
var vTrue = true

var (
	// PeriodicFullScan represents the global program used for periodic full scans.
	PeriodicFullScan = Program{
		ID:       "periodic-full-scan",
		Name:     "Periodic Scan",
		Disabled: &vFalse,
		Policies: []PolicyGroup{
			PolicyGroup{
				Group:  "default-global",
				Policy: "default-global",
			},
			PolicyGroup{
				Group:  "sensitive-global",
				Policy: "sensitive-global",
			},
		},
		DefaultMetadata: api.GlobalProgramsMetadata{
			// Minute | Hour | Dom | Month | Dow
			// Standard crontab specs, e.g. "* * * * ?"
			// Descriptors, e.g. "@midnight", "@every 1h30m"
			Cron: "0 8 * * 1", // Run the scan every Monday at 8am UTC.

			// Autosend is set by default to false for this program.
			Autosend: &vFalse,
		},
	}
	// RedconScan represents the global program used for periodic scans
	// of the Redcon discovered assets.
	RedconScan = Program{
		ID:       "redcon-scan",
		Name:     "Redcon Scan",
		Disabled: &vTrue,
		Policies: []PolicyGroup{
			PolicyGroup{
				Group:  "redcon-global",
				Policy: "default-global",
			},
		},
		DefaultMetadata: api.GlobalProgramsMetadata{
			// Minute | Hour | Dom | Month | Dow
			// Standard crontab specs, e.g. "* * * * ?"
			// Descriptors, e.g. "@midnight", "@every 1h30m"
			Cron: "0 8 7 10 *", // Run the scan every October 7th at 8am UTC.
		},
	}
	// WebScanning represents the global program used for web scans
	WebScanning = Program{
		ID:       "web-scanning",
		Name:     "Web Scanning",
		Disabled: &vFalse,
		Policies: []PolicyGroup{
			PolicyGroup{
				Group:  "web-scanning-global",
				Policy: "web-scanning-global",
			},
		},
		DefaultMetadata: api.GlobalProgramsMetadata{
			// Minute | Hour | Dom | Month | Dow
			// Standard crontab specs, e.g. "* * * * ?"
			// Descriptors, e.g. "@midnight", "@every 1h30m"
			Cron: "0 8 15 * *", // Run the scan every 15th of the month at 8am UTC.

			// Autosend is set by default to false for this program.
			Autosend: &vFalse,
		},
	}
)
