/*
Copyright 2021 Adevinta
*/

package store

import (
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/gorm"

	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func mustGetFixtureAssets(t *testing.T, s *Store) []*api.Asset {
	assets := []*api.Asset{}
	res := s.Conn.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Order("id", true).
		Find(&assets)
	if res.Error != nil {
		t.Fatalf("error reading fixture assets %+v", res.Error)
	}
	return assets
}

func TestAssetsReaderRead(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		t.Fatalf("error reading test fixtures: %+v", err)
	}
	store := testStore.(Store)
	defer store.Close()
	type fields struct {
		Total    int
		next     string
		pageSize int
		tx       *gorm.DB
		finished bool
		total    int
		lock     bool
	}
	tests := []struct {
		name          string
		readerCreator func() (*AssetsReader, error)
		want          []*api.Asset
		wantErr       error
	}{
		{
			name: "ReturnsFirstPage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				return &reader, nil
			},
			want: mustGetFixtureAssets(t, &store)[:7],
		},
		{
			name: "ReturnsSecondPage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				_, err = reader.Read()
				if err != nil {
					return nil, err
				}
				return &reader, nil
			},
			want: mustGetFixtureAssets(t, &store)[7:14],
		},
		{
			name: "ReturnsThirdPage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				_, err = reader.Read()
				if err != nil {
					return nil, err
				}
				_, err = reader.Read()
				if err != nil {
					return nil, err
				}
				return &reader, nil
			},
			want:    mustGetFixtureAssets(t, &store)[14:17],
			wantErr: ErrReadAssetsFinished,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := tt.readerCreator()
			if err != nil {
				t.Fatalf("error creating AssetsReader %v", err)
			}
			defer a.Close()
			got, err := a.Read()
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Errorf("AssetsReader.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAsset}, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("got != want. Diff: %s\n", diff)
			}
		})
	}
}
