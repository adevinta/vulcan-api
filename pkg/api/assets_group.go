/*
Copyright 2021 Adevinta
*/

package api

import (
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/common"
)

const (
	DiscoveredAssetsGroupName  = "security-team-discovered-assets"
	WebScanningAssetsGroupName = "web-scanning"
)

type AssetGroup struct {
	AssetID   string    `gorm:"primary_key;AUTO_INCREMENT" json:"asset_id" validate:"required"`
	Asset     *Asset    `json:"asset"` // This line is infered from column name "asset_id".
	GroupID   string    `gorm:"primary_key;AUTO_INCREMENT" json:"group_id" validate:"required"`
	Group     *Group    `json:"group"` // This line is infered from column name "group_id".
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Group struct {
	ID          string        `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	TeamID      string        `json:"team_id"`
	Team        *Team         `json:"team"` // This line is infered from column name "team_id".
	Name        string        `json:"name" validate:"required"`
	Options     string        `json:"options"`
	AssetGroup  []*AssetGroup `json:"asset_group"` // This line is infered from other tables.
	Description *string       `json:"description,omitempty"`
	CreatedAt   time.Time     `json:"-"`
	UpdatedAt   time.Time     `json:"-"`
}

// Overwrite gorm default pluralized table name convention
func (AssetGroup) TableName() string {
	return "asset_group"
}

type GroupResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Options     string  `json:"options"`
	AssetsCount *int    `json:"assets_count,omitempty"`
}

func (g Group) ToResponse() *GroupResponse {
	response := &GroupResponse{
		ID:          g.ID,
		Name:        g.Name,
		Options:     g.Options,
		Description: g.Description,
	}

	if g.AssetGroup != nil {
		size := len(g.AssetGroup)
		response.AssetsCount = &size
	}

	return response
}

type AssetsGroupResponse struct {
	Assets []AssetResponse `json:"assets"`
	Group  GroupResponse   `json:"group"`
}

type AssetGroupResponse struct {
	Asset AssetResponse `json:"asset"`
	Group GroupResponse `json:"group"`
}

func (ag AssetGroup) ToResponse() AssetGroupResponse {
	response := AssetGroupResponse{}
	if ag.Asset != nil {
		response.Asset = ag.Asset.ToResponse()
	}
	if ag.Group != nil {
		response.Group = *ag.Group.ToResponse()
	}
	return response
}

func (ag AssetGroup) Validate() error {
	if ag.Group != nil {
		err := ag.Group.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g Group) Validate() error {
	validationErr := validator.New().Struct(g)
	if validationErr != nil {
		return validationErr
	}
	if !common.IsStringEmpty(&g.Options) && !common.IsValidJSON(&g.Options) {
		return errors.Validation("group.options field identified by has invalid json")
	}
	return nil
}
