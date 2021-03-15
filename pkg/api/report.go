/*
Copyright 2021 Adevinta
*/

package api

import "time"

type Report struct {
	ID             string     `json:"id"`
	ScanID         string     `json:"scan_id"`
	ProgramName    string     `json:"program_name"`
	Report         string     `json:"report"`
	ReportJson     string     `json:"report_json"`
	EmailBody      string     `json:"email_body"`
	DeliveredTo    string     `json:"delivered_to"`
	UpdateStatusAt *time.Time `json:"update_status_at"`
	Status         string     `json:"status"`
	Risk           *int       `json:"risk"`
	CreatedAt      *time.Time `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
}

type ReportResponse struct {
	ReportID    string `json:"report_id"`
	ScanID      string `json:"scan_id"`
	ProgramName string `json:"program_name"`
	Report      string `json:"report"`
	ReportJson  string `json:"report_json"`
	Status      string `json:"status"`
	DeliveredTo string `json:"delivered_to"`
	Risk        *int   `json:"risk"`
}

func (r Report) ToResponse() *ReportResponse {
	response := ReportResponse{
		ReportID:    r.ID,
		ScanID:      r.ScanID,
		ProgramName: r.ProgramName,
		Report:      r.Report,
		ReportJson:  r.ReportJson,
		Status:      r.Status,
		DeliveredTo: r.DeliveredTo,
		Risk:        r.Risk,
	}
	return &response
}

type ReportEmailResponse struct {
	EmailBody string `json:"email_body"`
}

func (r Report) ToEmailResponse() *ReportEmailResponse {
	response := ReportEmailResponse{
		EmailBody: r.EmailBody,
	}
	return &response
}
