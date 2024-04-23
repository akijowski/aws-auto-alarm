package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog/log"
)

const (
	eventbridgeEventSource     = "aws.tag"
	eventbridgeEventDetailType = "Tag Change on Resource"
)

type AlarmCreator interface {
	Create(ctx context.Context, req *alarmCreationRequest) error
}

type AlarmHandler struct {
	alarmCreators map[string]AlarmCreator
}

type tagChangeDetail struct {
	ChangedTagKeys []string          `json:"changed-tag-keys"`
	Service        string            `json:"service"`
	ResourceType   string            `json:"resource-type"`
	Version        int64             `json:"version"`
	Tags           map[string]string `json:"tags"`
}

type alarmCreationRequest struct {
	tagChange *tagChangeDetail
	event     *events.EventBridgeEvent
}

func (h *AlarmHandler) Handle(ctx context.Context, event *events.SQSEvent) (*events.SQSEventResponse, error) {
	log.Ctx(ctx).Debug().Msg("handling sqs event")
	batchItemFailures := []events.SQSBatchItemFailure{}

	for _, record := range event.Records {
		rCtx := log.Ctx(ctx).With().Str("sqs_message_id", record.MessageId).Logger().WithContext(ctx)
		request, err := unmarshalRecord(record)
		if err != nil {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
			log.Ctx(rCtx).Error().Err(err).Msg("unable to unmarshal record")
			continue
		}

		if request.event.Source != eventbridgeEventSource || request.event.DetailType != eventbridgeEventDetailType {
			log.Ctx(rCtx).Warn().Msg("event does not match requirements, dropping message")
			continue
		}

		creator, ok := h.alarmCreators[request.tagChange.Service]
		if !ok {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
			log.Ctx(rCtx).Error().Err(err).Msg("alarm creator not found")
			continue
		}

		err = creator.Create(ctx, request)
		if err != nil {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
			log.Ctx(rCtx).Error().Err(err).Msg("unable to create alarm")
			continue
		}
	}

	return &events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
}

func unmarshalRecord(record events.SQSMessage) (*alarmCreationRequest, error) {
	var eventbridgeEvent *events.EventBridgeEvent
	err := json.Unmarshal([]byte(record.Body), &eventbridgeEvent)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal eventbridge event: %w", err)
	}

	var detail *tagChangeDetail
	err = json.Unmarshal(eventbridgeEvent.Detail, &detail)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal eventbridge event detail: %w", err)
	}

	return &alarmCreationRequest{
		tagChange: detail,
		event:     eventbridgeEvent,
	}, nil
}
