/*
Copyright 2021 Adevinta
*/

package api

import "time"

type Scan struct {
	ID            string     `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	ProgramID     string     `json:"program_id" validate:"required"`
	Program       *Program   `json:"program"`
	ScheduledTime *time.Time `json:"scheduled_time"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Progress      *float32   `json:"progress"`
	Status        string     `json:"status"`
	CheckCount    *int       `json:"check_count,omitempty"`
	RequestedBy   string     `json:"requested_by"`
	ReportLink    string     `json:"report_link"`
}

type ScanResponse struct {
	ID            string           `json:"id"`
	StartTime     *time.Time       `json:"start_time"`
	Endtime       *time.Time       `json:"end_time"`
	ScheduledTime *time.Time       `json:"scheduled_time"`
	Progress      *float32         `json:"progress"`
	CheckCount    *int             `json:"check_count,omitempty"`
	Status        string           `json:"status"`
	RequestedBy   string           `json:"requested_by"`
	ReportLink    string           `json:"report_link,omitempty"`
	Program       *ProgramResponse `json:"program"`
}

func (s Scan) ToResponse() *ScanResponse {
	response := ScanResponse{
		ID:            s.ID,
		StartTime:     s.StartTime,
		Endtime:       s.EndTime,
		ScheduledTime: s.ScheduledTime,
		Progress:      s.Progress,
		Status:        s.Status,
		RequestedBy:   s.RequestedBy,
		ReportLink:    s.ReportLink,
		CheckCount:    s.CheckCount,
	}
	if s.Program != nil {
		response.Program = s.Program.ToResponse()
	}
	return &response
}
