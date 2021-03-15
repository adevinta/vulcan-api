/*
Copyright 2021 Adevinta
*/

package api

// Healthcheck ....
type Healthcheck struct {
	Status string `json:"status" validate:"required"`
}

// ToResponse ...
func (h Healthcheck) ToResponse() HealthcheckResponse {
	return HealthcheckResponse(h)
}

// HealthcheckResponse ...
type HealthcheckResponse struct {
	Status string `json:"status"`
}
