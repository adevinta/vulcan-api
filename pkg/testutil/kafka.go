package testutil

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

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
	newTopics := map[string]string{}
	var newTopicNames []string
	for entity, topic := range topics {
		pc, _, _, _ := runtime.Caller(1)
		callerName := strings.Replace(runtime.FuncForPC(pc).Name(), ".", "_", -1)
		callerName = strings.Replace(callerName, "-", "_", -1)
		parts := strings.Split(callerName, "/")
		name := strings.ToLower(fmt.Sprintf("%s_%s_test", topic, parts[len(parts)-1]))
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

	var topics []kafka.TopicSpecification
	for _, name := range names {
		topic := kafka.TopicSpecification{
			Topic:         name,
			NumPartitions: 1,
		}
		topics = append(topics, topic)
	}
	for {
		results, err = client.CreateTopics(context.Background(), topics, opTimeout)
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
