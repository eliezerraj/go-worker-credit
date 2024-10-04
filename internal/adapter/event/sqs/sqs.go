package sqs

import (
	"os"
	"os/signal"
	"context"
	"syscall"
	"time"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/lib"
	"github.com/go-worker-credit/internal/service"
	"github.com/go-worker-credit/internal/config/config_aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var childLogger = log.With().Str("event", "sqs").Logger()

type NotifierSQS struct{
	sqsClient 	*sqs.Client
	queueConfig	*core.QueueConfig
	workerService	*service.WorkerService
}

func NewNotifierSQS(ctx context.Context, 
					queueConfig *core.QueueConfig,
					workerService	*service.WorkerService) (*NotifierSQS, error){
	childLogger.Debug().Msg("NewNotifierSQS")

	span := lib.Span(ctx, "event.NewNotifierSQS")	
    defer span.End()

	sdkConfig, err := config_aws.GetAWSConfig(ctx, queueConfig.AwsRegion)
	if err != nil{
		return nil, err
	}

	sqsClient := sqs.NewFromConfig(*sdkConfig)

	notifierSQS := NotifierSQS{
		sqsClient: sqsClient,
		queueConfig: queueConfig,
		workerService: 	workerService,
	}

	return &notifierSQS, nil
} 

func (s *NotifierSQS) Consumer(	ctx context.Context, 
								wg *sync.WaitGroup, 
								appServer core.WorkerAppServer ) {
	childLogger.Debug().Msg("Consumer")

	defer func() { 
		childLogger.Debug().Msg("Closing consumer waiting please !!!")
		defer wg.Done()
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	event := core.Event{}
	for run {
		select {
		case sig := <-sigchan:
			childLogger.Debug().Interface("Caught signal terminating: ", sig).Msg("")
			run = false
		case <-time.After(5 * time.Second): // Timeout for ending the consumer
			childLogger.Debug().Msg("No messages after timeout, stopping!!!")
            return
		default:
			receiveMsgInput := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(s.queueConfig.QueueUrl),
				MaxNumberOfMessages: 1,      // The number of messages to receive
				WaitTimeSeconds:     10,     // Long polling: Wait time to receive a message
				VisibilityTimeout:   30,     // Visibility timeout for processing the message
			}

			result, err := s.sqsClient.ReceiveMessage(ctx, receiveMsgInput)
			if err != nil {
				childLogger.Error().Err(err).Msg("error sqsClient.ReceiveMessageInput")
			}

			if len(result.Messages) == 0 {
                childLogger.Debug().Msg("Consumer No messages received")
                continue
            }

			for _, message := range result.Messages {
				childLogger.Debug().Interface("message.Body:",message.Body).Msg("")

				json.Unmarshal([]byte(*message.Body), &event)

				childLogger.Debug().Msg("++++++++++++++++++++++++++++++++++++++++++")
				childLogger.Debug().Interface(">>>>>> AccountIDTo:    ",event.EventData.Transfer.AccountIDTo).Msg("<<<<<<<")
				childLogger.Debug().Msg("++++++++++++++++++++++++++++++++++++++++++")

				err = s.workerService.CreditFundSchedule(ctx, *event.EventData.Transfer)
				if err != nil {
					childLogger.Error().Err(err).Msg("Erro no Consumer.CreditFundSchedule")
					childLogger.Debug().Msg("ROLLBACK!!!!")
				}
				
				deleteMsgInput := &sqs.DeleteMessageInput{
                    QueueUrl:      aws.String(s.queueConfig.QueueUrl),
                    ReceiptHandle: message.ReceiptHandle,
                }
				_, err := s.sqsClient.DeleteMessage(ctx, deleteMsgInput)
				if err != nil {
					childLogger.Error().Err(err).Msg("Erro to delete message from SQS")
					childLogger.Debug().Msg("ROLLBACK!!!!")
				}

				childLogger.Debug().Msg("DELETE MESSAGE COMPLETE !!!!")
			}
		}
	}
}