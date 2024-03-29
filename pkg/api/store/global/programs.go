/*
Copyright 2021 Adevinta
*/

package global

import "github.com/adevinta/vulcan-api/pkg/api"

func init() {
	registerProgram(PeriodicFullScan)
	registerProgram(RedconScan)
	registerProgram(WebScanning)
	registerProgram(CPScan)
}

var vFalse = false
var vTrue = true

var (
	// PeriodicFullScan represents the global program used for periodic full scans.
	PeriodicFullScan = Program{
		ID:   "periodic-full-scan",
		Name: "Periodic Scan",
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

			// Disabled is set by default to false for this program.
			Disabled: &vFalse,
		},
	}
	// RedconScan represents the global program used for periodic scans
	// of the Redcon discovered assets.
	RedconScan = Program{
		ID:   "redcon-scan",
		Name: "Redcon Scan",
		Policies: []PolicyGroup{
			PolicyGroup{
				Group:  "redcon-global",
				Policy: "redcon-global",
			},
		},
		DefaultMetadata: api.GlobalProgramsMetadata{
			// Minute | Hour | Dom | Month | Dow
			// Standard crontab specs, e.g. "* * * * ?"
			// Descriptors, e.g. "@midnight", "@every 1h30m"
			Cron: "0 12 * * 2", // Run the scan every Tuesday at 12pm UTC.

			// Autosend is set by default to false for this program.
			Autosend: &vFalse,

			// Disabled is set by default to true for this program.
			Disabled: &vFalse,
		},
	}
	// WebScanning represents the global program used for web scans
	WebScanning = Program{
		ID:   "web-scanning",
		Name: "Web Scanning",
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
			Cron: "0 8 * * 3", // Run the scan every Wednesday at 8am UTC.

			// Autosend is set by default to false for this program.
			Autosend: &vFalse,

			// Disabled is set by default to false for this program.
			Disabled: &vFalse,
		},
	}
	// CPScan represents the global program used for periodic scans
	// of the Common Platform discovered assets.
	CPScan = Program{
		ID:   "cp-scan",
		Name: "CP Scan",
		Policies: []PolicyGroup{
			PolicyGroup{
				Group:  "cp-global",
				Policy: "cp-global",
			},
		},
		DefaultMetadata: api.GlobalProgramsMetadata{
			// Minute | Hour | Dom | Month | Dow
			// Standard crontab specs, e.g. "* * * * ?"
			// Descriptors, e.g. "@midnight", "@every 1h30m"
			Cron: "0 6 * * 3", // Run the scan every Wednesday at 6am UTC.

			// Autosend is set by default to false for this program.
			Autosend: &vFalse,

			// Disabled is set by default to true for this program.
			Disabled: &vFalse,
		},
	}
)
