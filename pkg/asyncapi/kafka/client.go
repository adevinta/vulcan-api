/*
Copyright 2022 Adevinta
*/

package kafka

import (
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	// ErrUndefinedEntity is returned by the Push method of the [Client] when the
	// given entity name is unknown.
	ErrUndefinedEntity = errors.New("undefined entity")
	// ErrEmptyPayload is returned by the Push method of the [Client] when the
	// given payload is empty.
	ErrEmptyPayload = errors.New("payload can't be empty")
)

const (
	kafkaSecurityProtocol = "sasl_ssl"
	kafkaSaslMechanisms   = "SCRAM-SHA-256"
)

// Client implements an EventStreamClient using Kafka as the event stream
// system.
type Client struct {
	producer *kafka.Producer
	// Contains the mappings beetween the entity names and the corresponding
	// underlaying kafka topics.
	Topics map[string]string
}

// NewClient creates a new Kafka client connected to the a broker using the
// given credentials and setting the mapping between all the entities and their
// corresponding topics.
func NewClient(user string, password string, broker string, topics map[string]string) (Client, error) {
	config := kafka.ConfigMap{
		"bootstrap.servers": broker,
	}
	if password != "" {
		config.SetKey("security.protocol", kafkaSecurityProtocol)
		config.SetKey("sasl.mechanisms", kafkaSaslMechanisms)
		config.SetKey("sasl.username", user)
		config.SetKey("sasl.password", password)
	}
	p, err := kafka.NewProducer(&config)
	if err != nil {
		return Client{}, err
	}
	return Client{p, topics}, nil
}

// Push sends the payload of an entity, with the specified id, to corresponding
// topic according to the specified entity, using the kafka broker the client
// is connected to. The method waits until kafka confirms the message has been
// stored in the topic.
func (c *Client) Push(entity string, id string, payload []byte, metadata map[string][]byte) error {
	topic, ok := c.Topics[entity]
	if !ok {
		return ErrUndefinedEntity
	}
	delivered := make(chan kafka.Event)
	defer close(delivered)
	var headers []kafka.Header
	for k, v := range metadata {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: v,
		})
	}
	msg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:     []byte(id),
		Value:   []byte(payload),
		Headers: headers,
	}
	err := c.producer.Produce(&msg, delivered)
	if err != nil {
		return fmt.Errorf("error producing message: %w", err)
	}
	e := <-delivered
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("error delivering message: %w", m.TopicPartition.Error)
	}
	return nil
}
