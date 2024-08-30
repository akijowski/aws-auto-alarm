package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/command"
)

const (
	eventbridgeEventSource     = "aws.tag"
	eventbridgeEventDetailType = "Tag Change on Resource"
)

type AlarmHandler struct {
	MetricAPI   autoalarm.MetricAlarmAPI
	ResourceAPI autoalarm.GetResourcesAPI
}

func (h *AlarmHandler) Handle(ctx context.Context, event *events.SQSEvent) (*events.SQSEventResponse, error) {
	log := zerolog.Ctx(ctx).With().
		Int("sqs_messages_count", len(event.Records)).
		Logger()
	log.Info().Msg("Received SQS event")

	// do this for now, make better later
	for _, record := range event.Records {
		if err := h.handleSQSRecord(ctx, record); err != nil {
			log.Error().Str("sqs_message_id", record.MessageId).Err(err).Msg("Failed to process SQS record")
			return nil, err
		}
	}

	return nil, nil
}

func (h *AlarmHandler) handleSQSRecord(ctx context.Context, record events.SQSMessage) error {
	log := zerolog.Ctx(ctx).With().
		Str("message_id", record.MessageId).
		Logger()
	log.Info().Msg("Processing SQS record")

	event := new(events.EventBridgeEvent)
	if err := json.Unmarshal([]byte(record.Body), event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	log.Debug().Interface("event", event).Msg("Unmarshalled event")

	log.Info().Str("source", event.Source).Str("detail_type", event.DetailType).Msg("Received EventBridge event")
	if err := filterEvent(event); err != nil {
		return fmt.Errorf("unable to process event: %w", err)
	}

	return buildAndRun(ctx, h.MetricAPI, log, event)
}

func filterEvent(event *events.EventBridgeEvent) error {
	if event.Source != eventbridgeEventSource || event.DetailType != eventbridgeEventDetailType {
		return fmt.Errorf("event source %s and detail-type %s does not match expected values", event.Source, event.DetailType)
	}

	resourceARN, err := arn.Parse(event.Resources[0])
	if err != nil {
		return fmt.Errorf("unable to parse resource ARN: %w", err)
	}

	if !slices.Contains([]string{"sqs"}, resourceARN.Service) {
		return fmt.Errorf("resource service %s is not supported", resourceARN.Service)
	}

	return nil
}

func buildAndRun(ctx context.Context, api autoalarm.MetricAlarmAPI, wr io.Writer, event *events.EventBridgeEvent) error {
	log := zerolog.Ctx(ctx)
	config, err := autoalarm.NewLambdaConfig(ctx, event)
	if err != nil {
		return fmt.Errorf("unable to create config: %w", err)
	}
	log.Info().Interface("config", config).Msg("Created config")

	// do this better
	cmdFactory := command.DefaultFactory(ctx, config)
	var cmd autoalarm.Command
	if config.DryRun {
		cmd, err = cmdFactory.WithWriter(wr)(ctx, config.Delete)
	} else {
		cmd, err = cmdFactory.WithMetricAPI(api)(ctx, config.Delete)
	}
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	if err = cmd.Execute(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	log.Info().Msg("event handling complete")
	return nil
}
