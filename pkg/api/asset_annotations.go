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

func (an AssetAnnotation) Validate() error {
	err := validator.New().Struct(an)
	if err != nil {
		return errors.Validation(err)
	}
	return nil
}

type AssetAnnotations []*AssetAnnotation

type AssetAnnotationsMap map[string]string

type AssetAnnotationsResponse struct {
	Annotations AssetAnnotationsMap `json:"annotations"`
}

func (ans AssetAnnotations) ToMap() AssetAnnotationsMap {
	m := AssetAnnotationsMap{}
	for _, an := range ans {
		m[an.Key] = an.Value
	}
	return m
}

func (anm AssetAnnotationsMap) ToModel() AssetAnnotations {
	annotations := AssetAnnotations{}
	for k, v := range anm {
		annotations = append(annotations, &AssetAnnotation{
			Key:   k,
			Value: v,
		})
	}
	return annotations
}

func (ans AssetAnnotations) ToResponse() AssetAnnotationsResponse {
	return AssetAnnotationsResponse{Annotations: ans.ToMap()}
}
