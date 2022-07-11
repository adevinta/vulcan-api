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
	// ErrUndefinedEntity is returned by the Push method of the Client when the
	// given entity name is unknown by the client.
	ErrUndefinedEntity = errors.New("undefined entity")
	// ErrEmptyPayload is returned by the Push method of the Client when the
	// given payload is empty.
	ErrEmptyPayload = errors.New("payload can't be empty")
)

// Client implements an EventStreamClient using Kafka as the event stream
// system.
type Client struct {
	producer *kafka.Producer
	topics   map[string]string
}

// NewClient creates a new Kafka client connected to the a broker using the
// given credentials and setting the mapping between all the entities and their
// corresponding topics.
func NewClient(user string, password string, broker string, topics map[string]string) (Client, error) {
	config := kafka.ConfigMap{
		"bootstrap.servers": broker,
		"security.protocol": "sasl_ssl",
		"sasl.mechanisms":   "SCRAM-SHA-256",
		"sasl.username":     user,
		"sasl.password":     password,
	}
	p, err := kafka.NewProducer(&config)
	if err != nil {
		return Client{}, err
	}
	return Client{p, topics}, nil
}

// Push sends the payload of an entity, with the specified id, to corresponding
// topic according to the specified entity, using the kafka broker the client
// is connected to. The method waits until kafka confirm the message has been
// stored in the topic. The payload can not be empty.
func (c *Client) Push(entity string, id string, payload []byte) error {
	if len(payload) == 0 {
		return ErrEmptyPayload
	}
	topic, ok := c.topics[entity]
	if !ok {
		return ErrUndefinedEntity
	}
	delivered := make(chan kafka.Event)
	defer close(delivered)
	msg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(payload),
		Value: []byte(payload),
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
