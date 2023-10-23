/*
Copyright 2021 Adevinta
*/

package reports

// Config contains the configuration
// parameters for report related actions.
type Config struct {
	SNSARN      string `mapstructure:"sns_arn"`
	SNSEndpoint string `mapstructure:"sns_endpoint"`
	InsecureTLS bool   `mapstructure:"insecure_tls"`
	VulcanUIURL string `mapstructure:"vulcanui_url"`
}

// Notification represents
// a report notification.
type Notification struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Format  string `json:"format"`
}

type teamInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
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
	InfoFixed     int    `json:"info_fixed"`
	LowFixed      int    `json:"low_fixed"`
	MediumFixed   int    `json:"medium_fixed"`
	HighFixed     int    `json:"high_fixed"`
	CriticalFixed int    `json:"critical_fixed"`
}
