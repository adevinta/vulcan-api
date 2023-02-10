package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaTestBroker contains the address of the local broker used for tests.
const KafkaTestBroker = "localhost:29092"

// PrepareKafka creates a new empty topic in the local test Kafka server for
// each topic name present in topics maps. The name of the topics created will
// have the following shape:
// <calling_function_name>_<original_topic_name>_test. It returns a new map
// with the entities remapped to new topics created.
func PrepareKafka(topics map[string]string) (map[string]string, error) {
	// Generate a unique deterministic topic name for the caller of this function.
	tRef := time.Now().Unix()
	newTopics := map[string]string{}
	var newTopicNames []string
	for entity, topic := range topics {
		pc, _, _, _ := runtime.Caller(1)
		callerName := strings.Replace(runtime.FuncForPC(pc).Name(), ".", "_", -1)
		callerName = strings.Replace(callerName, "-", "_", -1)
		parts := strings.Split(callerName, "/")
		name := strings.ToLower(fmt.Sprintf("%s_%s_%d_test", topic, parts[len(parts)-1], tRef))
		newTopics[entity] = name
		newTopicNames = append(newTopicNames, name)
	}

	err := createTopics(newTopicNames)
	if err != nil {
		return nil, err
	}
	return newTopics, nil
}

func createTopics(names []string) error {
	config := kafka.ConfigMap{
		"bootstrap.servers": KafkaTestBroker,
	}
	client, err := kafka.NewAdminClient(&config)
	if err != nil {
		return err
	}

	waitDuration := time.Duration(time.Second * 60)
	opTimeout := kafka.SetAdminOperationTimeout(waitDuration)

	results, err := client.DeleteTopics(context.Background(), names, opTimeout)
	if err != nil {
		return err
	}
	tResults := topicsOpResult(results)
	if tResults.Error() != kafka.ErrNoError && tResults.Error() != kafka.ErrUnknownTopicOrPart {
		return fmt.Errorf("error deleting topic %s", tResults.Error())
	}

	for _, name := range names {
		topic := kafka.TopicSpecification{
			Topic:         name,
			NumPartitions: 1,
		}
		// Retry the create topic operation until in does not return an error
		// indicating that the topic already exits, at which point we have a
		// clean new topic created.
		for {
			results, err = client.CreateTopics(context.Background(), []kafka.TopicSpecification{topic}, opTimeout)
			if err != nil {
				return err
			}
			tResults = topicsOpResult(results)
			if tResults.Error() == kafka.ErrNoError {
				break
			}
			if tResults.Error() == kafka.ErrTopicAlreadyExists {
				continue
			}
			return fmt.Errorf("error creating topics: %s", tResults.Error())
		}
	}

	return nil
}

type topicsOpResult []kafka.TopicResult

func (t topicsOpResult) Error() kafka.ErrorCode {
	for _, res := range t {
		if res.Error.Code() != kafka.ErrNoError {
			return res.Error.Code()
		}
	}
	return kafka.ErrNoError
}

type AssetTopicData struct {
	Payload asyncapi.AssetPayload
	Headers map[string][]byte
}

func ReadAllAssetsTopic(topic string) ([]AssetTopicData, error) {
	broker := KafkaTestBroker
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

	var topicAssetsData []AssetTopicData
Loop:
	for ev := range c.Events() {
		switch e := ev.(type) {
		case *confluentKafka.Message:
			data := e.Value
			asset := asyncapi.AssetPayload{}
			// The data will be empty in case the event is a tombstone.
			if len(data) > 0 {
				err = json.Unmarshal(data, &asset)
				if err != nil {
					return nil, err
				}
			}
			headers := map[string][]byte{}
			for _, v := range e.Headers {
				headers[v.Key] = v.Value
			}
			topicData := AssetTopicData{
				Payload: asset,
				Headers: headers,
			}
			topicAssetsData = append(topicAssetsData, topicData)
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
			break Loop
		default:
			return nil, fmt.Errorf("received unexpected message %v", e)
		}
	}
	return topicAssetsData, nil
}
