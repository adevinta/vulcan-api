/*
Copyright 2022 Adevinta
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	gokitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/adevinta/vulcan-api/pkg/asyncapi/kafka"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

type nullLogger struct {
}

func (n nullLogger) Errorf(s string, params ...any) {
}

func (n nullLogger) Infof(s string, params ...any) {

}

func (n nullLogger) Debugf(s string, params ...any) {

}

func TestBump(t *testing.T) {
	topics := map[string]string{asyncapi.AssetsEntityName: "assets"}
	testTopics, err := testutil.PrepareKafka(topics)
	if err != nil {
		t.Fatalf("error creating test topics: %v", err)
	}

	kclient, err := kafka.NewClient("", "", testutil.KafkaTestBroker, testTopics)
	if err != nil {
		t.Fatal(err)
	}
	vulcanStore, err := testutil.PrepareDatabaseLocal("../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatalf("error creating test database %v", err)
	}
	defer vulcanStore.Close()
	testStore := vulcanStore.(store.Store)

	vulcan := asyncapi.NewVulcan(&kclient, nullLogger{})

	glogger := gokitlog.NewNopLogger()
	nullLogger := levelLogger{glogger}

	allAssets, err := readAllAssetsDB(testStore)
	if err != nil {
		t.Fatalf("error reading assets from DB")
	}

	wantAssets := DBAssetsToAsyncAssets(allAssets)

	if err := bump(vulcan, testStore, 5, nullLogger); err != nil {
		t.Fatalf("error bumping assets %v", err)
	}
	topic := kclient.Topics[asyncapi.AssetsEntityName]
	gotAssets, err := readAllAssetsTopic(topic)
	if err != nil {
		t.Fatalf("error reading assets from kafka %v", err)
	}

	sortOpts := cmpopts.SortSlices(func(a, b asyncapi.AssetPayload) bool {
		return strings.Compare(a.Id, b.Id) < 0
	})
	diff := cmp.Diff(wantAssets, gotAssets, sortOpts)
	if diff != "" {
		t.Fatalf("want!=got, diff: %s", diff)
	}
}

func TestNoAssets(t *testing.T) {
	topics := map[string]string{asyncapi.AssetsEntityName: "assets"}
	testTopics, err := testutil.PrepareKafka(topics)
	if err != nil {
		t.Fatalf("error creating test topics: %v", err)
	}

	kclient, err := kafka.NewClient("", "", testutil.KafkaTestBroker, testTopics)
	if err != nil {
		t.Fatal(err)
	}
	dbName := testutil.DBNameForFunc(1)
	dsn, err := testutil.CreateTestDatabase(dbName)
	if err != nil {
		t.Fatal(err)
	}

	glogger := gokitlog.NewNopLogger()
	nullLogger := levelLogger{glogger}

	testStore, err := store.NewStore("", dsn, glogger, false, map[string][]string{})
	if err != nil {
		t.Fatal(err)
	}
	// Ensure there are no assets in the DB.
	res := testStore.Conn.Exec("DELETE FROM assets")
	if res.Error != nil {
		t.Fatal(res.Error)
	}

	vulcan := asyncapi.NewVulcan(&kclient, nullLogger)

	if err := bump(vulcan, testStore, 5, nullLogger); err != nil {
		t.Fatalf("error bumping assets %v", err)
	}
	topic := kclient.Topics[asyncapi.AssetsEntityName]
	gotAssets, err := readAllAssetsTopic(topic)
	if err != nil {
		t.Fatalf("error reading assets from kafka %v", err)
	}

	var wantAssets []asyncapi.AssetPayload

	sortOpts := cmpopts.SortSlices(func(a, b asyncapi.AssetPayload) bool {
		return strings.Compare(a.Id, b.Id) < 0
	})

	diff := cmp.Diff(wantAssets, gotAssets, sortOpts)
	if diff != "" {
		t.Fatalf("want!=got, diff: %s", diff)
	}
}

func DBAssetsToAsyncAssets(dbAssets []*api.Asset) []asyncapi.AssetPayload {
	var assets []asyncapi.AssetPayload
	for _, asset := range dbAssets {
		aAsset := asyncapi.AssetPayload{
			Id: asset.ID,
			Team: &asyncapi.Team{
				Id:          asset.Team.ID,
				Name:        asset.Team.Name,
				Description: asset.Team.Description,
				Tag:         asset.Team.Tag,
			},
			Alias:      asset.Alias,
			Rolfp:      asset.ROLFP.String(),
			Scannable:  *asset.Scannable,
			AssetType:  (*asyncapi.AssetType)(&asset.AssetType.Name),
			Identifier: asset.Identifier,
		}
		annotations := []*asyncapi.Annotation{}
		for _, a := range asset.AssetAnnotations {
			aAnnotation := &asyncapi.Annotation{
				Key:   a.Key,
				Value: a.Value,
			}
			annotations = append(annotations, aAnnotation)
		}
		aAsset.Annotations = annotations
		assets = append(assets, aAsset)
	}
	return assets

}

func readAllAssetsDB(s store.Store) ([]*api.Asset, error) {
	assets := make([]*api.Asset, 0)
	res := s.Conn.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Find(&assets)
	if res.Error != nil {
		return nil, res.Error
	}
	return assets, nil
}

func readAllAssetsTopic(topic string) ([]asyncapi.AssetPayload, error) {
	broker := testutil.KafkaTestBroker
	config := confluentKafka.ConfigMap{
		"go.events.channel.enable": true,
		"bootstrap.servers":        broker,
		"group.id":                 "test_" + topic,
		"enable.partition.eof":     true,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
	}
	c, err := confluentKafka.NewConsumer(&config)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	if err = c.Subscribe(topic, nil); err != nil {
		return nil, err
	}

	var assets []asyncapi.AssetPayload
LOOP:
	for ev := range c.Events() {
		switch e := ev.(type) {
		case *confluentKafka.Message:
			data := e.Value
			asset := asyncapi.AssetPayload{}
			err = json.Unmarshal(data, &asset)
			if err != nil {
				return nil, err
			}
			assets = append(assets, asset)
			_, err := c.CommitOffsets([]confluentKafka.TopicPartition{
				{
					Topic:     e.TopicPartition.Topic,
					Partition: e.TopicPartition.Partition,
					Offset:    e.TopicPartition.Offset + 1,
				},
			})
			if err != nil {
				return nil, err
			}
		case confluentKafka.Error:
			return nil, e
		case confluentKafka.PartitionEOF:
			break LOOP
		default:
			return nil, fmt.Errorf("received unexpected message %v", e)
		}
	}
	return assets, nil
}
