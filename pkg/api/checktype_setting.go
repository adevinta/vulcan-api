/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/common"
)

type ChecktypeSetting struct {
	ID            string     `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	PolicyID      string     `json:"policy_id"`
	Policy        *Policy    `json:"policy"` // This line is infered from column name "policy_id".
	CheckTypeName string     `json:"checktype_name"`
	Options       *string    `json:"options"`
	CreatedAt     *time.Time `json:"-"`
	UpdatedAt     *time.Time `json:"-"`
}

type ChecktypeSettingResponse struct {
	ID            string `json:"id"`
	CheckTypeName string `json:"checktype_name"`
	Options       string `json:"options"`
}

func (c ChecktypeSetting) ToResponse() *ChecktypeSettingResponse {
	response := ChecktypeSettingResponse{
		ID:            c.ID,
		CheckTypeName: c.CheckTypeName,
		Options:       common.StringValue(c.Options),
	}
	return &response
}

func (c ChecktypeSetting) Validate() error {
	if common.IsStringEmpty(&c.CheckTypeName) ||
		(!common.IsStringEmpty(c.Options) && !common.IsValidJSON(c.Options)) {
		return errors.Validation("Validation error invalid checktypeSetting payload")
	}
	return nil
}
