/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// GlobalPolicyConfig defines the global policy configuration
// in terms of checks and assettypes to process.
type GlobalPolicyConfig map[string]GlobalPolicyConfigEntry

// GlobalPolicyConfigEntry defines for a specific global policy
// the list of (allowed and blocked) (checks and assettypes) and
// a list of suffix to exclude if check name matches.
// Blocking takes precedence.
// Empty allowed slices means ALL allowed.
type GlobalPolicyConfigEntry struct {
	AllowedChecks     []string `mapstructure:"allowed_checks"`
	BlockedChecks     []string `mapstructure:"blocked_checks"`
	AllowedAssettypes []string `mapstructure:"allowed_assettypes"`
	BlockedAssettypes []string `mapstructure:"blocked_assettypes"`
	ExcludingSuffixes []string `mapstructure:"excluding_suffixes"`
}

func init() {
	registerPolicy(&DefaultPolicy{})
	registerPolicy(&SensitivePolicy{})
	registerPolicy(&WebScanningPolicy{})
	registerPolicy(&RedconPolicy{})
	registerPolicy(&CPPolicy{})
}

// DefaultPolicy contains all checks execpts the ones for docker images.
type DefaultPolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the policy.
func (d *DefaultPolicy) Name() string {
	return "default-global"
}

// Description returns a meanfull explanation of the group.
func (d *DefaultPolicy) Description() string {
	return "Default set of checktypes that will be executed against the assets present in the default-global group"
}

func (d *DefaultPolicy) Init(informer ChecktypesInformer) error {
	d.checktypeInformer = informer
	return nil
}

func (d *DefaultPolicy) Eval(ctx context.Context, gpc GlobalPolicyConfig) ([]*api.ChecktypeSetting, error) {
	checkTypesInfo, err := d.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
	}
	if config, ok := gpc[d.Name()]; ok {
		return evalWithConfig(ctx, config, checkTypesInfo)
	}
	// No need to use a map for a only a few element but it makes clear what
	// assettypes and check types are excluded.
	excludedAssettypes := map[string]struct{}{
		"": struct{}{},
	}
	excludedChecktypes := map[string]struct{}{
		"vulcan-masscan":              struct{}{},
		"vulcan-exposed-services":     struct{}{},
		"vulcan-exposed-router-ports": struct{}{},
		"vulcan-csp-report-uri":       struct{}{},
		"vulcan-docker-image":         struct{}{},
		"vulcan-zap":                  struct{}{},
		"vulcan-seekret":              struct{}{},
		"vulcan-retirejs":             struct{}{},
		"vulcan-tls":                  struct{}{},
		"vulcan-burp":                 struct{}{},
	}

	checktypes := []*api.ChecktypeSetting{}
	added := map[string]struct{}{}
	for a, c := range checkTypesInfo {
		if _, ok := excludedAssettypes[a]; ok {
			continue
		}
		for _, name := range c {
			if _, ok := excludedChecktypes[name]; ok {
				continue
			}
			// Exclude experimental checks.
			if strings.HasSuffix(name, "-experimental") {
				continue
			}
			// Checktypes can be repeated so only add a check if it was not
			// added previously.
			if _, ok := added[name]; ok {
				continue
			}
			checktypes = append(checktypes, &api.ChecktypeSetting{
				ID:            name,
				CheckTypeName: name,
			})
			added[name] = struct{}{}
		}
	}
	return checktypes, nil
}

type SensitivePolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the group.
func (d *SensitivePolicy) Name() string {
	return "sensitive-global"
}

// Description returns a meanfull explanation of the group.
func (d *SensitivePolicy) Description() string {
	return "Default set of checktypes that will be executed against the assets present in the sensitive-global group"
}

func (d *SensitivePolicy) Init(informer ChecktypesInformer) error {
	d.checktypeInformer = informer
	return nil
}

// Eval return same checktypes as default-global policy except vulcan-nessus.
func (d *SensitivePolicy) Eval(ctx context.Context, gpc GlobalPolicyConfig) ([]*api.ChecktypeSetting, error) {
	checkTypesInfo, err := d.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
	}
	if config, ok := gpc[d.Name()]; ok {
		return evalWithConfig(ctx, config, checkTypesInfo)
	}
	// The sensitive policy is the same that the default policy
	// but excluding Nessus.
	dp := &DefaultPolicy{checktypeInformer: d.checktypeInformer}
	checktypes, err := dp.Eval(ctx, gpc)
	if err != nil {
		return nil, err
	}
	index := -1
	for i, c := range checktypes {
		if c.CheckTypeName == "vulcan-nessus" {
			index = i
		}
	}
	if index == -1 {
		return checktypes, nil
	}
	return append(checktypes[:index], checktypes[index+1:]...), nil
}

// WebScanningPolicy contains all checks related with web scanning.
type WebScanningPolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the group.
func (ws *WebScanningPolicy) Name() string {
	return "web-scanning-global"
}

// Description returns a meanfull explanation of the group.
func (ws *WebScanningPolicy) Description() string {
	return "Default set of checktypes related with web scanning"
}

func (ws *WebScanningPolicy) Init(informer ChecktypesInformer) error {
	ws.checktypeInformer = informer
	return nil
}

func (ws *WebScanningPolicy) Eval(ctx context.Context, gpc GlobalPolicyConfig) ([]*api.ChecktypeSetting, error) {
	checkTypesInfo, err := ws.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
	}
	if config, ok := gpc[ws.Name()]; ok {
		return evalWithConfig(ctx, config, checkTypesInfo)
	}

	webScanningChecktypes := map[string]struct{}{
		"vulcan-zap": struct{}{},
	}

	checktypes := []*api.ChecktypeSetting{}
	added := map[string]struct{}{}
	for _, c := range checkTypesInfo {
		for _, name := range c {
			if _, ok := webScanningChecktypes[name]; !ok {
				continue
			}
			// Exclude experimental checks.
			if strings.HasSuffix(name, "-experimental") {
				continue
			}
			// Checktypes can be repeated so only add a check if it was not
			// added previously.
			if _, ok := added[name]; ok {
				continue
			}
			checktypes = append(checktypes, &api.ChecktypeSetting{
				ID:            name,
				CheckTypeName: name,
			})
			added[name] = struct{}{}
		}
	}
	return checktypes, nil
}

// RedconPolicy contains all checks associated with the "DefaultPolicy", but
// excluding "vulcan-nessus"
type RedconPolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the group.
func (r *RedconPolicy) Name() string {
	return "redcon-global"
}

// Description returns a meaningful explanation of the group.
func (r *RedconPolicy) Description() string {
	return "Default set of checktypes that will be executed against the assets present in the redcon-global group"
}

func (r *RedconPolicy) Init(informer ChecktypesInformer) error {
	r.checktypeInformer = informer
	return nil
}

func (r *RedconPolicy) Eval(ctx context.Context, gpc GlobalPolicyConfig) ([]*api.ChecktypeSetting, error) {
	checkTypesInfo, err := r.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
	}
	if config, ok := gpc[r.Name()]; ok {
		return evalWithConfig(ctx, config, checkTypesInfo)
	}

	// The Redcon policy is the same that the Default policy
	// but excluding "vulcan-nessus".
	dp := &DefaultPolicy{checktypeInformer: r.checktypeInformer}
	checktypes, err := dp.Eval(ctx, gpc)
	if err != nil {
		return nil, err
	}
	index := -1
	for i, c := range checktypes {
		if c.CheckTypeName == "vulcan-nessus" {
			index = i
		}
	}
	if index == -1 {
		return checktypes, nil
	}
	return append(checktypes[:index], checktypes[index+1:]...), nil
}

// CPPolicy contains all checks associated with the "DefaultPolicy", but
// excluding "vulcan-nessus"
type CPPolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the group.
func (r *CPPolicy) Name() string {
	return "cp-global"
}

// Description returns a meaningful explanation of the group.
func (r *CPPolicy) Description() string {
	return "Default set of checktypes that will be executed against the assets present in the cp-global group"
}

func (r *CPPolicy) Init(informer ChecktypesInformer) error {
	r.checktypeInformer = informer
	return nil
}

func (r *CPPolicy) Eval(ctx context.Context, gpc GlobalPolicyConfig) ([]*api.ChecktypeSetting, error) {
	checkTypesInfo, err := r.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
	}
	if config, ok := gpc[r.Name()]; ok {
		return evalWithConfig(ctx, config, checkTypesInfo)
	}

	// The CP policy is the same that the Default policy
	// but excluding "vulcan-nessus".
	dp := &DefaultPolicy{checktypeInformer: r.checktypeInformer}
	checktypes, err := dp.Eval(ctx, gpc)
	if err != nil {
		return nil, err
	}
	index := -1
	for i, c := range checktypes {
		if c.CheckTypeName == "vulcan-nessus" {
			index = i
		}
	}
	if index == -1 {
		return checktypes, nil
	}
	return append(checktypes[:index], checktypes[index+1:]...), nil
}

func evalWithConfig(ctx context.Context, config GlobalPolicyConfigEntry, cti map[string][]string) ([]*api.ChecktypeSetting, error) {
	checktypes := []*api.ChecktypeSetting{}
	added := map[string]struct{}{}
	for at, c := range cti {
		// If allowed assettypes list is not empty and the assetype is not included in the allowed list then skip.
		// If the assettype is in blocked list then skip.
		if (len(config.AllowedAssettypes) > 0 && !inSlice(at, config.AllowedAssettypes)) || inSlice(at, config.BlockedAssettypes) {
			continue
		}
		for _, checkName := range c {
			// If allowed check list is not empty and the check is not included in the allowed list then skip.
			// If the check is in the blocked list then skip.
			if (len(config.AllowedChecks) > 0 && !inSlice(checkName, config.AllowedChecks)) || inSlice(checkName, config.BlockedChecks) {
				continue
			}
			// If the check name contains an excluding suffix pattern then skip.
			if len(config.ExcludingSuffixes) > 0 {
				exclude := false
				for _, suffix := range config.ExcludingSuffixes {
					if strings.HasSuffix(checkName, suffix) {
						exclude = true
						break
					}
				}
				if exclude {
					continue
				}
			}
			// Checktypes can be repeated so only add a check if it was not
			// added previously.
			if _, ok := added[checkName]; ok {
				continue
			}
			checktypes = append(checktypes, &api.ChecktypeSetting{
				ID:            checkName,
				CheckTypeName: checkName,
			})
			added[checkName] = struct{}{}
		}
	}
	return checktypes, nil
}

func inSlice(item string, s []string) bool {
	for _, v := range s {
		if v == item {
			return true
		}
	}
	return false
}
