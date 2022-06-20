/*
Copyright 2022 Adevinta
*/

package store

import (
	"log"
	"testing"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestStoreCreateAssetAnnotations(t *testing.T) {
	db, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	tests := []struct {
		name        string
		teamID      string
		assetID     string
		annotations []*api.AssetAnnotation
		want        []*api.AssetAnnotation
		wantErr     error
	}{
		{
			name:    "CreatesNonExistentAssetsAnnotations",
			teamID:  "ea686be5-be9b-473b-ab1b-621a4f575d51",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "new",
					Value: "value",
				},
			},
			want: []*api.AssetAnnotation{
				{
					Key:   "new",
					Value: "value",
				},
			},
		},
		{
			name:    "DontCreateAnnotationsIfAlreadyExist",
			teamID:  "ea686be5-be9b-473b-ab1b-621a4f575d51",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "autodiscovery/security/keytoupdate",
					Value: "updated",
				},
			},
			wantErr: errors.Create("annotation 'autodiscovery/security/keytoupdate' already present for asset id '73e33dcb-d07c-41d1-bc32-80861b49941e'"),
		},
		{
			name:    "DontCreateAssetsAnnotationsForInvalidTeam",
			teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "new",
					Value: "value",
				},
			},
			wantErr: errors.Forbidden("asset does not belong to team"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.CreateAssetAnnotations(tt.teamID, tt.assetID, tt.annotations)
			if errToStr(err) != errToStr(tt.wantErr) {

				t.Fatalf("got error != want err, %+v!=%+v", errToStr(err), errToStr(tt.wantErr))
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("want annotations != got annotations, diff:\n %s", diff)
			}
		})
	}
}

func TestStoreUpdateAssetAnnotations(t *testing.T) {
	db, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	tests := []struct {
		name        string
		teamID      string
		assetID     string
		annotations []*api.AssetAnnotation
		want        []*api.AssetAnnotation
		wantErr     error
	}{
		{
			name:    "DontUpdateAssetsAnnotationsForInvalidTeam",
			teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "keywithoutprefix",
					Value: "value",
				},
			},
			wantErr: errors.Forbidden("asset does not belong to team"),
		},
		{
			name:    "UpdatesExistentAssetsAnnotations",
			teamID:  "ea686be5-be9b-473b-ab1b-621a4f575d51",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "autodiscovery/security/keytoupdate",
					Value: "updatedvalue",
				},
			},
			want: []*api.AssetAnnotation{
				{
					Key:   "autodiscovery/security/keytoupdate",
					Value: "updatedvalue",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.UpdateAssetAnnotations(tt.teamID, tt.assetID, tt.annotations)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("got error != want err, %+v!=%+v", errToStr(err), errToStr(tt.wantErr))
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("want annotations != got annotations, diff:\n %s", diff)
			}
		})
	}
}

func TestStorePutAssetAnnotations(t *testing.T) {
	db, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	tests := []struct {
		name        string
		teamID      string
		assetID     string
		annotations []*api.AssetAnnotation
		want        []*api.AssetAnnotation
		wantErr     error
	}{
		{
			name:    "DontPutAssetsAnnotationsForInvalidTeam",
			teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "newkey",
					Value: "newvalue",
				},
			},
			wantErr: errors.Forbidden("asset does not belong to team"),
		},
		{
			name:    "ReplaceExistentAssetsAnnotations",
			teamID:  "ea686be5-be9b-473b-ab1b-621a4f575d51",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "newkey1",
					Value: "newvalue1",
				},
				{
					Key:   "newkey2",
					Value: "newvalue2",
				},
			},
			want: []*api.AssetAnnotation{
				{
					Key:   "newkey1",
					Value: "newvalue1",
				},
				{
					Key:   "newkey2",
					Value: "newvalue2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.PutAssetAnnotations(tt.teamID, tt.assetID, tt.annotations)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("got error != want err, %+v!=%+v", errToStr(err), errToStr(tt.wantErr))
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("want annotations != got annotations, diff:\n %s", diff)
			}
		})
	}
}

func TestStoreDeleteAssetAnnotations(t *testing.T) {
	db, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	tests := []struct {
		name        string
		teamID      string
		assetID     string
		annotations []*api.AssetAnnotation
		want        []*api.AssetAnnotation
		wantErr     error
	}{
		{
			name:    "DontDeleteAssetsAnnotationsForInvalidTeam",
			teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key:   "newkey",
					Value: "newvalue",
				},
			},
			wantErr: errors.Forbidden("asset does not belong to team"),
		},
		{
			name:    "DeleteAssetsAnnotations",
			teamID:  "ea686be5-be9b-473b-ab1b-621a4f575d51",
			assetID: "73e33dcb-d07c-41d1-bc32-80861b49941e",
			annotations: []*api.AssetAnnotation{
				{
					Key: "autodiscovery/security/keytodelete",
				},
			},
			want: []*api.AssetAnnotation{
				{
					Key:   "keywithoutprefix",
					Value: "valuewithoutprefix",
				},
				{
					Key:   "autodiscovery/security/keytoupdate",
					Value: "valuetoupdate",
				},
				{
					Key:   "autodiscovery/security/keytonotupdate",
					Value: "valuetonotupdate",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.DeleteAssetAnnotations(tt.teamID, tt.assetID, tt.annotations)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("got error != want err, %+v!=%+v", errToStr(err), errToStr(tt.wantErr))
			}
			got, err := db.ListAssetAnnotations(tt.teamID, tt.assetID)
			if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
				t.Fatal(err)
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Fatalf("want annotations != got annotations, diff:\n %s", diff)
			}
		})
	}
}
