/*
Copyright 2021 Adevinta
*/

package api

import "strings"

type AssetType struct {
	ID     string   `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	Name   string   `json:"name"`
	Assets []*Asset `json:"assets"` // This line is infered from other tables.
}

func (at AssetType) ToResponse() AssetTypeResponse {
	return AssetTypeResponse{
		ID:   at.ID,
		Name: at.Name,
	}
}

type AssetTypeResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ValidAssetType indicates if the asset type name exists in Vulcan.
func ValidAssetType(assetTypeName string) bool {
	valid := []string{"AWSAccount", "IP", "IPRange", "DomainName", "Hostname", "DockerImage", "WebAddress", "GitRepository", "GCPProject"}
	for _, a := range valid {
		if strings.EqualFold(a, assetTypeName) {
			return true
		}
	}
	return false
}
