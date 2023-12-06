/*
Copyright 2021 Adevinta
*/

package api

import (
	"database/sql/driver"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/common"
	types "github.com/adevinta/vulcan-types"
)

// DiscoveredAssetsGroupSuffix is used by the Merge Discovered Assets feature
// to restrict the discovery onboarding to Groups with a name containing that
// suffix.
const DiscoveredAssetsGroupSuffix = "-discovered-assets"

var ErrROLFPInvalidText = "invalid ROLFP representation"

type Asset struct {
	ID                string             `gorm:"primary_key;AUTO_INCREMENT" json:"id" sql:"DEFAULT:gen_random_uuid()"`
	TeamID            string             `json:"team_id" validate:"required"`
	Team              *Team              `json:"team,omitempty"` // This line is infered from column name "team_id".
	AssetTypeID       string             `json:"asset_type_id" validate:"required"`
	AssetType         *AssetType         `json:"asset_type"` // This line is infered from column name "asset_type_id".
	Identifier        string             `json:"identifier" validate:"required"`
	Alias             string             `json:"alias"`
	Options           *string            `json:"options"`
	EnvironmentalCVSS *string            `json:"environmental_cvss"`
	ROLFP             *ROLFP             `json:"rolfp" sql:"DEFAULT:'R:1/O:1/L:1/F:1/P:1+S:2'"`
	Scannable         *bool              `json:"scannable" gorm:"default:true"`
	AssetGroups       []*AssetGroup      `json:"groups"`      // This line is infered from other tables.
	AssetAnnotations  []*AssetAnnotation `json:"annotations"` // This line is infered from other tables.
	CreatedAt         time.Time          `json:"-"`
	UpdatedAt         time.Time          `json:"-"`
	ClassifiedAt      *time.Time         `json:"classified_at"`
}

func validateAWSARN(arn string) bool {
	// This is a regular expression that matches AWS ARNs.
	arnRegex, err := regexp.Compile(`^arn:(aws|aws-cn|aws-us-gov):([a-z0-9-]+):([a-z\d-]*):([0-9]*):([a-zA-Z0-9_-]*)(//?[a-zA-Z0-9_-]+)*(//.*)?.*$`)
	if err != nil {
		// Return false if there has been an error compiling the string.
		return false
	}
	// Check if the ARN matches the regular expression.
	return arnRegex.MatchString(arn)
}

// Validate checks if an asset is valid.
func (a Asset) Validate() error {
	err := validator.New().Struct(a)
	if err != nil {
		return errors.Validation(err)
	}
	if !common.IsStringEmpty(a.Options) && !common.IsValidJSON(a.Options) {
		return errors.Validation("asset.options field has invalid json")
	}

	switch a.AssetType.Name {
	case "Hostname":
		if os.Getenv("VULCAN_HOSTNAME_VALIDATION_WITH_DNS") == "false" {
			if !types.IsHostnameNoDnsResolution(a.Identifier) {
				return errors.Validation("Identifier is not a valid Hostname")
			}
		} else {
			if !types.IsHostname(a.Identifier) {
				return errors.Validation("Identifier is not a valid Hostname")
			}
		}
	case "AWSAccount":
		if !validateAWSARN(a.Identifier) || !types.IsAWSARN(a.Identifier) {
			return errors.Validation("Identifier is not a valid AWSAccount")
		}
	case "GCPProject":
		if !types.IsGCPProjectID(a.Identifier) {
			return errors.Validation("Identifier is not a valid GCPProject")
		}
	case "DockerImage":
		if !types.IsDockerImage(a.Identifier) {
			return errors.Validation("Identifier is not a valid DockerImage")
		}
	case "GitRepository":
		if !types.IsGitRepository(a.Identifier) {
			return errors.Validation("Identifier is not a valid GitRepository")
		}
	case "IP":
		if strings.HasSuffix(a.Identifier, "/32") {
			if !types.IsHost(a.Identifier) {
				return errors.Validation("Identifier is not a valid Host")
			}
		} else {
			if !types.IsIP(a.Identifier) {
				return errors.Validation("Identifier is not a valid IP")
			}
		}
	case "IPRange":
		if !types.IsCIDR(a.Identifier) {
			return errors.Validation("Identifier is not a valid CIDR block")
		}
	case "WebAddress":
		if !types.IsWebAddress(a.Identifier) {
			return errors.Validation("Identifier is not a valid WebAddress")
		}
	case "DomainName":
		if ok, _ := types.IsDomainName(a.Identifier); !ok {
			return errors.Validation("Identifier is not a valid DomainName")
		}
	default:
		// If none of the previous case match, force a validation error
		return errors.Validation("Asset type not supported")
	}

	return nil
}

func (a Asset) ToResponse() AssetResponse {
	assetReponse := AssetResponse{}
	if a.AssetType != nil {
		assetReponse.AssetType = a.AssetType.ToResponse()
	}
	assetReponse.ID = a.ID
	assetReponse.Identifier = a.Identifier
	assetReponse.Options = a.Options
	assetReponse.EnvironmentalCVSS = a.EnvironmentalCVSS
	assetReponse.ROLFP = a.ROLFP
	assetReponse.Scannable = a.Scannable
	assetReponse.ClassifiedAt = a.ClassifiedAt
	assetReponse.Alias = a.Alias

	if a.AssetGroups != nil {
		for _, ag := range a.AssetGroups {
			if ag.Group != nil {
				assetReponse.Groups = append(assetReponse.Groups, ag.Group.ToResponse())
			}
		}
	}

	if len(a.AssetAnnotations) > 0 {
		var ans AssetAnnotations = a.AssetAnnotations
		assetReponse.Annotations = ans.ToMap()
	}

	return assetReponse
}

// Validate validates the values stored in the receiver are in the specified
// range: 0 to 1 for Reputation, Operation, Legal, Financial and Personal 0.
// range: 0 to 2 for Scope.
func (r *ROLFP) Validate() error {
	if r.IsEmpty {
		return nil
	}
	if r.Reputation > 1 {
		return fmt.Errorf("invalid ROLFP field reputation value %d", r.Reputation)
	}

	if r.Operation > 1 {
		return fmt.Errorf("invalid ROLFP field operation value %d", r.Operation)
	}

	if r.Legal > 1 {
		return fmt.Errorf("invalid ROLFP field legal value %d", r.Legal)
	}

	if r.Financial > 1 {
		return fmt.Errorf("invalid ROLFP field financial value %d", r.Financial)
	}

	if r.Personal > 1 {
		return fmt.Errorf("invalid ROLFP reputation value %d", r.Reputation)
	}

	if r.Scope > 2 {
		return fmt.Errorf("invalid ROLFP scope %d", r.Scope)
	}

	return nil
}

// Value returns the value of the ROLFP encoded to be persisted as a string.
func (r *ROLFP) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	val, err := r.MarshalText()
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (r *ROLFP) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Ivalid type decoding ROLFP from store %+v", value)
	}
	return r.UnmarshalText([]byte(str))
}

var DefaultROLFP = &ROLFP{
	Reputation: 1,
	Operation:  1,
	Legal:      1,
	Financial:  1,
	Personal:   1,
	Scope:      2,
	IsEmpty:    false,
}

type AssetResponse struct {
	ID                string              `json:"id"`
	AssetType         AssetTypeResponse   `json:"type"` // This line is infered from column name "asset_type_id".
	Identifier        string              `json:"identifier"`
	Alias             string              `json:"alias"`
	Options           *string             `json:"options"`
	EnvironmentalCVSS *string             `json:"environmental_cvss"`
	ROLFP             *ROLFP              `json:"rolfp"`
	Scannable         *bool               `json:"scannable"`
	ClassifiedAt      *time.Time          `json:"classified_at"`
	Groups            []*GroupResponse    `json:"groups"`
	Annotations       AssetAnnotationsMap `json:"annotations"`
}

type AssetCreationResponse struct {
	ID                string            `json:"id,omitempty"`
	Identifier        string            `json:"identifier"`
	AssetType         AssetTypeResponse `json:"type"` // This line is infered from column name "asset_type_id".
	Alias             string            `json:"alias"`
	Options           *string           `json:"options"`
	EnvironmentalCVSS *string           `json:"environmental_cvss"`
	ROLFP             *ROLFP            `json:"rolfp"`
	Scannable         *bool             `json:"scannable"`
	ClassifiedAt      *time.Time        `json:"classified_at"`
	Status            interface{}       `json:"status,omitempty"`
}

type Status struct {
	Code int `json:"code"`
}

// AssetMergeOperations defines a set of operations to perform when merging a
// list of assets requested by a discovery service.
type AssetMergeOperations struct {
	// Create assets that didn't exist yet in the team.
	Create []Asset
	// Associate already existing asset to the discovery group.
	Assoc []Asset
	// Update assets that were already existing (e.g. the scannable field or
	// the annotations)
	Update []Asset
	// Deassociate assets that haven't been discovered in the current discovery
	// operation, but that belong to other groups.
	Deassoc []Asset
	// Delete assets that haven't been discovered in the current discovery
	// operation and do not belong to other groups.
	Del []Asset

	// The team where the operations will be performed.
	TeamID string
	// The discovery group.
	Group Group
}
