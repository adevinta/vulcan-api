/*
Copyright 2021 Adevinta
*/

package api

import (
	"strings"
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

// Matches returns true if the current object exactly matches (both key and value)
// the asset annotation map passed as parameter. If a prefix is specified, only
// the keys matching the prefix are evaluated
func (ans AssetAnnotationsMap) Matches(annotations AssetAnnotationsMap, prefix string) bool {
	// Search a into b
	for k, v := range ans {
		if strings.HasPrefix(k, prefix) {
			if value, ok := annotations[k]; !ok || v != value {
				return false
			}
		}
	}

	// Search b into a
	for k, v := range annotations {
		if strings.HasPrefix(k, prefix) {
			if value, ok := ans[k]; !ok || v != value {
				return false
			}
		}
	}

	return true
}

// Merge takes an annotation map as input and merges it into the "base" annotation
// map, giving priority to the values of the former.
// If a prefix is specified, elements from the "base" map whose keys match the
// prefix are discarded
func (ans AssetAnnotationsMap) Merge(annotations AssetAnnotationsMap, prefix string) AssetAnnotationsMap {
	// Make a copy of the input annotations
	output := AssetAnnotationsMap{}
	for k, v := range annotations {
		output[k] = v
	}

	for k, v := range ans {
		if _, ok := output[k]; ok {
			continue
		} else if !ok && prefix == "" {
			output[k] = v
		} else if !ok && prefix != "" && !strings.HasPrefix(k, prefix) {
			output[k] = v
		}
	}

	return output
}
