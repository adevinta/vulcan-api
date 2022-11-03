/*
Copyright 2022 Adevinta
*/

package kafka

import (
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestClient_Push(t *testing.T) {
	topics := map[string]string{"assets": "assets"}
	testTopics, err := testutil.PrepareKafka(topics)
	if err != nil {
		t.Fatalf("error creating test Kafka client: %v", err)
	}
	client, err := NewClient("", "", testutil.KafkaTestBroker, testTopics)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		client   Client
		entity   string
		id       string
		payload  []byte
		metadata map[string][]byte
		want     []kafka.Message
		wantErr  bool
	}{
		{
			name:    "PushesAssetsToKafka",
			client:  client,
			entity:  "assets",
			payload: []byte("payload"),
			metadata: map[string][]byte{
				"key1": []byte("value"),
			},
			id: "id1",
			want: []kafka.Message{
				{
					Key:   []byte("id1"),
					Value: []byte("payload"),
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.client
			if err := c.Push(tt.entity, tt.id, tt.payload, tt.metadata); (err != nil) != tt.wantErr {
				t.Errorf("Client.Push() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := readAllTopic(testTopics["assets"])
			if err != nil {
				t.Fatal(err)
			}
			ignoreFields := cmpopts.IgnoreFields(kafka.Message{}, "TopicPartition", "Timestamp", "TimestampType", "Opaque")
			diff := cmp.Diff(tt.want, got, ignoreFields)
			if diff != "" {
				t.Fatalf("want!=got, diff: %s", diff)
			}

		})
	}
}

func readAllTopic(topic string) ([]kafka.Message, error) {
	broker := testutil.KafkaTestBroker
	config := kafka.ConfigMap{
		"go.events.channel.enable": true,
		"bootstrap.servers":        broker,
		"group.id":                 "test",
		"enable.partition.eof":     true,
		"auto.offset.reset":        "earliest",
	}
	c, err := kafka.NewConsumer(&config)
	if err != nil {
		return nil, err
	}
	if err = c.Subscribe(topic, nil); err != nil {
		return nil, err
	}

	var msgs []kafka.Message
LOOP:
	for ev := range c.Events() {
		switch e := ev.(type) {
		case *kafka.Message:

			msgs = append(msgs, *e)
			_, err := c.CommitOffsets([]kafka.TopicPartition{
				{
					Topic:     e.TopicPartition.Topic,
					Partition: e.TopicPartition.Partition,
					Offset:    e.TopicPartition.Offset + 1,
				},
			})
			if err != nil {
				return nil, err
			}
		case kafka.PartitionEOF:
			break LOOP
		default:
			return nil, fmt.Errorf("received unexpected message %v", e)
		}
	}
	return msgs, nil
}
