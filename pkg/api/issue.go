/*
Copyright 2024 Adevinta
*/

package api

type Issue struct {
	ID          string `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	CWEID       int    `json:"cwe_id" gorm:"Column:cwe_id"`
}

func (i Issue) ToResponse() *IssueResponse {
	return &IssueResponse{
		ID:          i.ID,
		Summary:     i.Summary,
		Description: i.Description,
		CWEID:       i.CWEID,
	}
}

type IssueResponse struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	CWEID       int    `json:"cwe_id"`
}
