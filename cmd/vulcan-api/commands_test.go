/*
Copyright 2021 Adevinta
*/

package main

import (
	"reflect"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api/store/global"
)

func initGlobalPolicyConfig(policyName string, settings map[string][]string) global.GlobalPolicyConfig {
	gpc := global.GlobalPolicyConfig{}
	var gpce global.GlobalPolicyConfigEntry
	for k, v := range settings {
		switch k {
		case "AllowedAssettypes":
			gpce.AllowedAssettypes = v
		case "BlockedAssettypes":
			gpce.BlockedAssettypes = v
		case "AllowedChecks":
			gpce.AllowedChecks = v
		case "BlockedChecks":
			gpce.BlockedChecks = v
		case "ExcludingSuffixes":
			gpce.ExcludingSuffixes = v
		}
	}
	if gpce.AllowedAssettypes == nil {
		gpce.AllowedAssettypes = []string{}
	}
	if gpce.BlockedAssettypes == nil {
		gpce.BlockedAssettypes = []string{}
	}
	if gpce.AllowedChecks == nil {
		gpce.AllowedChecks = []string{}
	}
	if gpce.BlockedChecks == nil {
		gpce.BlockedChecks = []string{}
	}
	if gpce.ExcludingSuffixes == nil {
		gpce.ExcludingSuffixes = []string{}
	}
	gpc[policyName] = gpce
	return gpc
}

func TestInitConfigGlobalPolicy(t *testing.T) {
	tests := []struct {
		name                   string
		cfgFile                string
		wantGlobalPolicyConfig global.GlobalPolicyConfig
	}{
		{
			name:                   "Empty",
			cfgFile:                "testdata/globalpolicy/empty.toml",
			wantGlobalPolicyConfig: nil,
		},
		{
			name:                   "Happy",
			cfgFile:                "testdata/globalpolicy/happy.toml",
			wantGlobalPolicyConfig: initGlobalPolicyConfig("happy", map[string][]string{}),
		},
		{
			name:    "Custom1",
			cfgFile: "testdata/globalpolicy/custom1.toml",
			wantGlobalPolicyConfig: initGlobalPolicyConfig("custom1", map[string][]string{
				"AllowedChecks":     {"check1", "check2"},
				"AllowedAssettypes": {"assettypeA"},
				"ExcludingSuffixes": {"-experimental"},
			}),
		},
		{
			name:    "Custom1",
			cfgFile: "testdata/globalpolicy/all-settings.toml",
			wantGlobalPolicyConfig: initGlobalPolicyConfig("all-settings", map[string][]string{
				"AllowedChecks":     {"check1"},
				"BlockedChecks":     {"check2"},
				"AllowedAssettypes": {"assettypeA"},
				"BlockedAssettypes": {"assettypeB"},
				"ExcludingSuffixes": {"-experimental"},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile = tt.cfgFile
			cfg = config{}
			initConfig()
			if !reflect.DeepEqual(cfg.GlobalPolicyConfig, tt.wantGlobalPolicyConfig) {
				t.Errorf("unexpeced global policy config parsing:\ngot=\n%#v\n, want=\n%#v", cfg.GlobalPolicyConfig, tt.wantGlobalPolicyConfig)
			}
		})
	}
}
