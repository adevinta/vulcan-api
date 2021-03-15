/*
Copyright 2021 Adevinta
*/

package queue

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	// MaxNumberOfMessages: The maximum number of messages to return.
	// Amazon SQS never returns more messages than this value (however,
	// fewer messages might be returned).
	// Valid values: 1 to 10. Default: 1.
	maxNumberOfMsg = 10

	// maxNumberOfMssgReads: The maximum number of times a message can
	// be read and parsed by the processor without success before being
	// deleted from the queue.
	maxNumberOfMssgReads = 25
)

var (
	// errInvalidMssg indicates that the queue message is invalid
	// or has a bad format.
	errInvalidMssg = errors.New("Invalid message")

	// errErrorReadCanceled is the error code that will be returned by an
	// SQS API request that was canceled. Requests given a aws.Context may
	// return this error when canceled.
	errErrorReadCanceled = errors.New("Canceled read from sqs")
)

// Config holds the required sqs config information.
type Config struct {
	QueueArn  string `mapstructure:"queue_arn"`
	QueueName string `mapstructure:"queue_name"`
	Endpoint  string `mapstructure:"endpoint"`
	WaitTime  int64  `mapstructure:"wait_time"`
	Timeout   int64  `mapstructure:"timeout"`
	Enabled   bool   `mapstructure:"enabled"`
}

// The message's contents (not URL-encoded).
type message struct {
	Message *string
}

// MessageProcessor process a message.
type MessageProcessor func(context.Context, []byte) error

// SQSConsumer reads and consumes sqs messages.
type SQSConsumer struct {
	sqs           sqsiface.SQSAPI
	processor     MessageProcessor
	logger        log.Logger
	receiveParams sqs.ReceiveMessageInput
	enabled       bool
}

// NewConsumer creates and initializes an SQSConsumer
func NewConsumer(c Config, log log.Logger, processor MessageProcessor) (*SQSConsumer, error) {
	if !c.Enabled {
		return nil, nil
	}

	sqsConsumer := &SQSConsumer{
		enabled:   c.Enabled,
		processor: processor,
		logger:    log,
	}

	return setupAWSParamsForSQSConsumer(c, log, sqsConsumer)
}

func setupAWSParamsForSQSConsumer(c Config, log log.Logger, sqsConsumer *SQSConsumer) (*SQSConsumer, error) {
	// get a new AWS session
	sess, err := session.NewSession()
	if err != nil {
		_ = level.Error(log).Log("CreatingAWSSession", err)
		return nil, err
	}

	arn, err := arn.Parse(c.QueueArn)
	if err != nil {
		return nil, err
	}
	awsCfg := aws.NewConfig()
	if arn.Region != "" {
		awsCfg = awsCfg.WithRegion(arn.Region)
	}
	if c.Endpoint != "" {
		awsCfg = awsCfg.WithEndpoint(c.Endpoint)
	}
	sqsConsumer.sqs = sqs.New(sess, awsCfg)

	input := sqs.GetQueueUrlInput{
		QueueName: aws.String(arn.Resource),
	}
	if arn.AccountID != "" {
		input.SetQueueOwnerAWSAccountId(arn.AccountID)
	}

	// get the SQS Queue URL
	resp, err := sqsConsumer.sqs.GetQueueUrl(&input)
	if err != nil {
		_ = level.Error(log).Log("ErrorRetrievingSQSURL", err)
		return nil, err
	}

	// configure SQS Receive Parameters
	sqsConsumer.receiveParams = sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(*resp.QueueUrl),
		MaxNumberOfMessages: aws.Int64(maxNumberOfMsg),
		WaitTimeSeconds:     aws.Int64(c.WaitTime),
		VisibilityTimeout:   aws.Int64(c.Timeout),
		AttributeNames:      []*string{aws.String(sqs.MessageSystemAttributeNameApproximateReceiveCount)},
	}

	// returns the SQS consumer
	return sqsConsumer, nil
}

// ProcessMessages reads messages from a SQS queue and sends
// them to the MessageProcessor.
func (s *SQSConsumer) ProcessMessages(ctx context.Context) {
	if !s.enabled {
		_ = level.Info(s.logger).Log("SQSConsumerEnabled", s.enabled)
		return
	}

	var exit bool
	var wg sync.WaitGroup

	for !exit {
		select {
		case <-ctx.Done():
			wg.Wait()
			exit = true
		default:
			err := s.processMessages(ctx, &wg)
			if err != nil {
				_ = level.Error(s.logger).Log("ErrorReadingMessages", err)
			}
		}
	}
}

// processMessages receives messages from an SQS queue, iterates over them and
// deletes them if message has been processed succesfully.
func (s *SQSConsumer) processMessages(ctx context.Context, wg *sync.WaitGroup) error {
	messages, err := s.receiveMessages(ctx)
	if err != nil {
		return err
	}

	for _, m := range messages {
		wg.Add(1)

		go func(ctx context.Context, wg *sync.WaitGroup, m *sqs.Message) {
			// mssgRcvdCount holds the number of times
			// the message has been readed from the queue
			// and not deleted.
			mssgRcvdCount, err := strconv.Atoi(*m.Attributes[sqs.MessageSystemAttributeNameApproximateReceiveCount])
			if err != nil {
				_ = level.Error(s.logger).Log("ErrorParsingMssgReceivedCount")
			}

			err = s.processMessage(ctx, m)

			// If there was an error processing mssg which is not
			// an errInvalidMssg error and the number of times
			// the mssg has been readed from queue is < to max,
			// return so message is not removed from the queue.
			if err != nil && err != errInvalidMssg && mssgRcvdCount < maxNumberOfMssgReads {
				_ = level.Error(s.logger).Log("ErrorProcessingMessage", err)
				return
			}
			_ = level.Debug(s.logger).Log("MsgProcessed", *m.MessageId)

			// Delete message from queue.
			if err := s.deleteMessage(ctx, m); err != nil {
				_ = level.Error(s.logger).Log("ErrorDeletingMessage", err)
			} else {
				_ = level.Debug(s.logger).Log("MsgDeleted", *m.MessageId)
			}

			wg.Done()
		}(ctx, wg, m)
	}
	return nil
}

// receiveMessages reads a SQS queue and returns the available messages
func (s *SQSConsumer) receiveMessages(ctx context.Context) ([]*sqs.Message, error) {
	resp, err := s.sqs.ReceiveMessageWithContext(ctx, &s.receiveParams)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			return nil, errErrorReadCanceled
		}
		return nil, err
	}
	return resp.Messages, nil
}

// processMessage validates the message and sends it to the custom processor
func (s *SQSConsumer) processMessage(ctx context.Context, m *sqs.Message) error {
	if m == nil {
		_ = level.Error(s.logger).Log("GotNilSQSMessage")
		return errInvalidMssg
	}

	if m.Body == nil {
		_ = level.Error(s.logger).Log("SQSMessageWithoutBody", m)
		return errInvalidMssg
	}

	err := s.customProcessor(ctx, *m)
	if err != nil {
		_ = level.Error(s.logger).Log("ErrorProcessingSqsMessage", err.Error())
		return err
	}

	return nil
}

// process Unmarshal a SQS message and invokes the custom processor
func (s *SQSConsumer) customProcessor(ctx context.Context, m sqs.Message) error {
	_ = level.Info(s.logger).Log("ProcessingMessageWithID", *m.MessageId)

	e := message{}

	err := json.Unmarshal([]byte(*m.Body), &e)
	if err != nil {
		return errInvalidMssg
	}

	if e.Message == nil {
		return errInvalidMssg
	}

	return s.processor(ctx, []byte(*e.Message))
}

// deleteMessage deletes a SQS message from the queue
func (s *SQSConsumer) deleteMessage(ctx context.Context, m *sqs.Message) error {
	_, err := s.sqs.DeleteMessage(&sqs.DeleteMessageInput{
		ReceiptHandle: m.ReceiptHandle,
		QueueUrl:      s.receiveParams.QueueUrl,
	})
	if err != nil {
		_ = level.Error(s.logger).Log("ErrorDeletingSQSMessage", err)
		return err
	}
	return nil
}
