/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"

	"github.com/adevinta/errors"
	"gopkg.in/go-playground/validator.v9"
)

type AssetAnnotation struct {
	AssetID   string    `gorm:"primary_key" json:"asset_id" validate:"required"`
	Asset     *Asset    `json:"asset"` // This line is infered from column name "asset_id".
	Key       string    `gorm:"primary_key" json:"key" validate:"required"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type AssetAnnotationResponse map[string]string

func (an AssetAnnotation) Validate() error {
	err := validator.New().Struct(an)
	if err != nil {
		return errors.Validation(err)
	}
	return nil
}
