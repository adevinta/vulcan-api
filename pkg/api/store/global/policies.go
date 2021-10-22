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

func init() {
	registerPolicy(&DefaultPolicy{})
	registerPolicy(&SensitivePolicy{})
	registerPolicy(&WebScanningPolicy{})
	registerPolicy(&RedconPolicy{})
}

// DefaultPolicy contains all checks execpts the ones for docker images.
type DefaultPolicy struct {
	checktypeInformer ChecktypesInformer
}

// Name returns the name of the group.
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

func (d *DefaultPolicy) Eval(ctx context.Context) ([]*api.ChecktypeSetting, error) {
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
	}

	checkTypesInfo, err := d.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
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

// Eval return all the checktypes except the ones executed against the docker
// images.
func (d *SensitivePolicy) Eval(ctx context.Context) ([]*api.ChecktypeSetting, error) {
	// The sensitive policy is the same that the default policy
	// but excluding Nessus.
	dp := &DefaultPolicy{checktypeInformer: d.checktypeInformer}
	checktypes, err := dp.Eval(ctx)
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

// WebScanningPolicy contains all checks related with web scannong.
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

func (ws *WebScanningPolicy) Eval(ctx context.Context) ([]*api.ChecktypeSetting, error) {
	webScanningChecktypes := map[string]struct{}{
		"vulcan-zap": struct{}{},
	}

	checkTypesInfo, err := ws.checktypeInformer.ByAssettype(ctx)
	if err != nil {
		return nil, errors.Default(err)
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

func (r *RedconPolicy) Eval(ctx context.Context) ([]*api.ChecktypeSetting, error) {
	// The Redcon policy is the same that the Default policy
	// but excluding "vulcan-nessus".
	dp := &DefaultPolicy{checktypeInformer: r.checktypeInformer}
	checktypes, err := dp.Eval(ctx)
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
