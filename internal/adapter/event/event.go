package event

import (
	go_core_event "github.com/eliezerraj/go-core/event/kafka"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("adapter", "event").Logger()

var consumerWorker go_core_event.ConsumerWorker

type WorkerEvent struct {
	Topics	[]string
	WorkerKafka *go_core_event.ConsumerWorker 
}

func NewWorkerEvent(topics []string, kafkaConfigurations *go_core_event.KafkaConfigurations) (*WorkerEvent, error) {
	childLogger.Info().Msg("NewWorkerEvent")

	workerKafka, err := consumerWorker.NewConsumerWorker(kafkaConfigurations)
	if err != nil {
		return nil, err
	}

	return &WorkerEvent{
		Topics: topics,
		WorkerKafka: workerKafka,
	},nil
}