/*
Copyright 2022 Adevinta
*/

package store

import (
	"errors"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/google/go-cmp/cmp"

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
	tests := []struct {
		name          string
		readerCreator func() (*AssetsReader, error)
		want          bool
		wantAssets    []*api.Asset
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
			want:       true,
			wantAssets: mustGetFixtureAssets(t, &store)[:7],
		},
		{
			name: "ReturnsSecondPage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				reader.Read()
				return &reader, nil
			},
			want:       true,
			wantAssets: mustGetFixtureAssets(t, &store)[7:14],
		},
		{
			name: "ReturnsThirdPage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				reader.Read()
				reader.Read()
				return &reader, nil
			},
			wantAssets: mustGetFixtureAssets(t, &store)[14:18],
			want:       true,
		},
		{
			name: "ReturnsNoMoreAssets",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 7)
				if err != nil {
					return nil, err
				}
				reader.Read()
				reader.Read()
				reader.Read()
				return &reader, nil
			},
			wantAssets: nil,
			want:       false,
		},
		{
			name: "ReturnsAssetsWhenLimitMultNOfAssets",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 17)
				if err != nil {
					return nil, err
				}
				reader.Read()
				return &reader, nil
			},
			wantAssets: mustGetFixtureAssets(t, &store)[17:18],
			want:       true,
		},
		{
			name: "ReturnsAllAssetsInOnePage",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 20)
				if err != nil {
					return nil, err
				}
				return &reader, nil
			},
			wantAssets: mustGetFixtureAssets(t, &store),
			want:       true,
		},
		{
			name: "ReturnsError",
			readerCreator: func() (*AssetsReader, error) {
				reader, err := store.NewAssetReader(true, 20)
				if err != nil {
					return nil, err
				}
				reader.tx.AddError(errors.New("simulated error"))
				return &reader, nil
			},
			wantAssets: mustGetFixtureAssets(t, &store),
			want:       false,
			wantErr:    errors.New("error reading assets: simulated error"),
		},
		{
			name: "ReturnsNoAssets",
			readerCreator: func() (*AssetsReader, error) {
				err := store.Conn.Exec("DELETE FROM assets").Error
				if err != nil {
					return nil, err
				}

				reader, err := store.NewAssetReader(true, 20)
				if err != nil {
					return nil, err
				}
				return &reader, nil
			},
			wantAssets: nil,
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := tt.readerCreator()
			if err != nil {
				t.Fatalf("error creating AssetsReader %v", err)
			}
			defer a.Close()
			got := a.Read()
			if got != tt.want {
				t.Errorf("go != want, got: %v, want: %v", got, tt.want)
				return
			}
			err = a.Err()
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Errorf("got error != want err, got: %v, want: %v", err, tt.wantErr)
				return
			}
			// If we got false as return value in the last call to Read, the
			// reader does not guarantee anything about the assets.
			if got == false {
				return
			}
			gotAssets := a.Assets()
			diff := cmp.Diff(tt.wantAssets, gotAssets, cmp.Options{ignoreFieldsAsset}, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("got assets != want assets. Diff: %s\n", diff)
			}
		})
	}
}
