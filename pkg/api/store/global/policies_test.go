/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/adevinta/vulcan-api/pkg/api"
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
			got, err := d.Eval(context.Background())
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
