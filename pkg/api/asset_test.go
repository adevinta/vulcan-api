/*
Copyright 2021 Adevinta
*/

package api

import (
	"errors"
	"testing"
)

func errToStr(err error) string {
	result := ""
	if err != nil {
		result = err.Error()
	}
	return result
}

func TestROLFP_String(t *testing.T) {
	tests := []struct {
		name  string
		ROLFP *ROLFP
		want  string
	}{
		{
			name: "should return empty string if IsEmpty is true",
			ROLFP: &ROLFP{
				IsEmpty: true,
			},
			want: "",
		},
		{
			name: "should return properly convert to string when all the fields are set",
			ROLFP: &ROLFP{
				Reputation: 1,
				Operation:  1,
				Legal:      1,
				Financial:  1,
				Personal:   1,
				Scope:      2,
			},
			want: "R:1/O:1/L:1/F:1/P:1+S:2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.ROLFP.String(); got != tt.want {
				t.Errorf("ROLFP.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssetValidate(t *testing.T) {
	tests := []struct {
		name    string
		asset   Asset
		wantErr error
	}{
		{
			name: "Correct AWS ARN",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "AWSAccount"},
				Identifier:  "arn:aws:iam::012345678900:root",
			},
			wantErr: nil,
		},
		{
			name: "Correct Complete AWS ARN",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "AWSAccount"},
				Identifier:  "arn:aws:iam:us-east-1:123456789012:user/Development/product_1234/*",
			},
		},
		{
			name: "Malformed AWS ARN with whitespaces",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "AWSAccount"},
				Identifier:  "arn:aws:iam:: 012345678900:root",
			},
			wantErr: errors.New("Identifier is not a valid AWSAccount"),
		},
		{
			name: "Correct GCP Project ID",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "GCPProject"},
				Identifier:  "bagase-crucible-bubble-gorilla",
			},
			wantErr: nil,
		},
		{
			name: "Malformed GCP Project ID ending with hyphen",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "GCPProject"},
				Identifier:  "inherent-derris-",
			},
			wantErr: errors.New("Identifier is not a valid GCPProject"),
		},
		{
			name: "Malformed GCP Project ID starting with numbers",
			asset: Asset{
				TeamID:      "TeamID",
				AssetTypeID: "AssetTypeID",
				AssetType:   &AssetType{Name: "GCPProject"},
				Identifier:  "007bond",
			},
			wantErr: errors.New("Identifier is not a valid GCPProject"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate(false)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
		})
	}
}
