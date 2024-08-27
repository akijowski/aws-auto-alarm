package task

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

const (
	eventbridgeEventSource     = "aws.tag"
	eventbridgeEventDetailType = "Tag Change on Resource"
)

type AlarmHandler struct {
}

type tagChangeDetail struct {
	ChangedTagKeys []string          `json:"changed-tag-keys"`
	Service        string            `json:"service"`
	ResourceType   string            `json:"resource-type"`
	Version        int64             `json:"version"`
	Tags           map[string]string `json:"tags"`
}

func (h *AlarmHandler) Handle(ctx context.Context, event *events.SQSEvent) (*events.SQSEventResponse, error) {
	log := zerolog.Ctx(ctx).With().
		Int("sqs_messages_count", len(event.Records)).
		Logger()
	log.Info().Msg("Received SQS event")

	return nil, nil
}
