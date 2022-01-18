/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/google/go-cmp/cmp"
)

type inMemoryChecktypesInformer struct {
	checktypes map[string][]string
}

func (i *inMemoryChecktypesInformer) ByAssettype(ctx context.Context) (map[string][]string, error) {
	return i.checktypes, nil
}

func TestDefaultPolicy_Eval(t *testing.T) {
	tests := []struct {
		name              string
		checktypeInformer ChecktypesInformer
		want              []*api.ChecktypeSetting
		wantErr           bool
	}{
		{
			name: "ExcludesAndDedupsChecks",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"DomainName": []string{"includeCheck", "vulcan-masscan", "vulcan-exposed-services", "vulcan-exposed-router-ports"},
					"IP":         []string{"includeCheck"},
				},
			},
			want: []*api.ChecktypeSetting{&api.ChecktypeSetting{ID: "includeCheck", CheckTypeName: "includeCheck"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultPolicy{
				checktypeInformer: tt.checktypeInformer,
			}
			var gpc GlobalPolicyConfig
			got, err := d.Eval(context.Background(), gpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultPolicy.Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestEvalWithConfig(t *testing.T) {
	tests := []struct {
		name              string
		checktypeInformer ChecktypesInformer
		gpc               GlobalPolicyConfig
		want              []*api.ChecktypeSetting
		wantErr           bool
	}{
		{
			name: "HappyPath",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check1", CheckTypeName: "check1"},
				{ID: "check2", CheckTypeName: "check2"},
				{ID: "check3", CheckTypeName: "check3"},
			},
		},
		{
			name: "AllowedChecksOnly",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					AllowedChecks: []string{"check1", "check2"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check1", CheckTypeName: "check1"},
				{ID: "check2", CheckTypeName: "check2"},
			},
		},
		{
			name: "BlockedChecksOnly",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					BlockedChecks: []string{"check1", "check2"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check3", CheckTypeName: "check3"},
			},
		},
		{
			name: "AllowedAndBlockedChecks",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					AllowedChecks: []string{"check1", "check2"},
					BlockedChecks: []string{"check1"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check2", CheckTypeName: "check2"},
			},
		},
		{
			name: "AllowedAssettypes",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					AllowedAssettypes: []string{"AssetType2"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check1", CheckTypeName: "check1"},
			},
		},
		{
			name: "ExcludingSufixes",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check1"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					ExcludingSuffixes: []string{"2", "3"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check1", CheckTypeName: "check1"},
			},
		},
		{
			name: "AllOptionsScenario",
			checktypeInformer: &inMemoryChecktypesInformer{
				checktypes: map[string][]string{
					"AssetType1": {"check1", "check2", "check3"},
					"AssetType2": {"check2"},
					"AssetType3": {"check4"},
				},
			},
			gpc: map[string]GlobalPolicyConfigEntry{
				"default-global": {
					AllowedAssettypes: []string{"AssetType1", "AssetType2"},
					BlockedAssettypes: []string{"AssetType3"},
					AllowedChecks:     []string{"check1", "check2", "check3", "check4"},
					BlockedChecks:     []string{"check3"},
					ExcludingSuffixes: []string{"2"},
				},
			},
			want: []*api.ChecktypeSetting{
				{ID: "check1", CheckTypeName: "check1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultPolicy{
				checktypeInformer: tt.checktypeInformer,
			}
			got, err := d.Eval(context.Background(), tt.gpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultPolicy.Eval() (with config) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}
