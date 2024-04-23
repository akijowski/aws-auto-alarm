package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
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
	batchItemFailures := []events.SQSBatchItemFailure{}

	for _, record := range event.Records {
		request, err := unmarshalRecord(record)
		if err != nil {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
			continue
		}

		if request.event.Source != eventbridgeEventSource || request.event.DetailType != eventbridgeEventDetailType {
			// log
			continue
		}

		creator, ok := h.alarmCreators[request.tagChange.Service]
		if !ok {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
			continue
		}

		err = creator.Create(ctx, request)
		if err != nil {
			failure := events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
			batchItemFailures = append(batchItemFailures, failure)
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
