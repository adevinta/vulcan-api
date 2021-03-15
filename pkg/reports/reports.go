/*
Copyright 2021 Adevinta
*/

package reports

import "time"

// Config contains the configuration
// parameters for report related actions.
type Config struct {
	SNSARN          string `mapstructure:"sns_arn"`
	SNSEndpoint     string `mapstructure:"sns_endpint"`
	APIBaseURL      string `mapstructure:"api_base_url"`
	InsecureTLS     bool   `mapstructure:"insecure_tls"`
	ScanRedirectURL string `mapstructure:"scan_redirect_url"`
	VulcanUIURL     string `mapstructure:"vulcanui_url"`
}

// ScanReport represents a scan report
// data as returned by reports generator API.
type ScanReport struct {
	ID            string    `json:"id"`
	ReportURL     string    `json:"report_url"`
	ReportJSONURL string    `json:"report_json_url"`
	ScanID        string    `json:"scan_id"`
	ProgramName   string    `json:"program_name"`
	Status        string    `json:"status"`
	Risk          int       `json:"risk"`
	EmailSubject  string    `json:"email_subject"`
	EmailBody     string    `json:"email_body"`
	DeliveredTo   string    `json:"delivered_to"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Notification represents
// a report notification.
type Notification struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Format  string `json:"format"`
}

// genReportEvent represents the
// payload for a report generation event.
type genReportEvent struct {
	Typ      string   `json:"type"`
	TeamInfo teamInfo `json:"team_info"`
	Data     scanData `json:"data"`
	AutoSend bool     `json:"auto_send"`
}
type teamInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
}
type scanData struct {
	ScanID      string `json:"scan_id"`
	ProgramName string `json:"program_name"`
}

// genDigestReportEvent represents the
// payload for a digest report generation event.
type genDigestReportEvent struct {
	Typ      string           `json:"type"`
	TeamInfo teamInfo         `json:"team_info"`
	Data     digestReportData `json:"data"`
	AutoSend bool             `json:"auto_send"`
}

type digestReportData struct {
	TeamID        string `json:"team_id"`
	DateFrom      string `json:"date_from"`
	DateTo        string `json:"date_to"`
	LiveReportURL string `json:"live_report_url"`
	Info          int    `json:"info"`
	Low           int    `json:"low"`
	Medium        int    `json:"medium"`
	High          int    `json:"high"`
	Critical      int    `json:"critical"`
	InfoDiff      int    `json:"info_diff"`
	LowDiff       int    `json:"low_diff"`
	MediumDiff    int    `json:"medium_diff"`
	HighDiff      int    `json:"high_diff"`
	CriticalDiff  int    `json:"critical_diff"`
}
