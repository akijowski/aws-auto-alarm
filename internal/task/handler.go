package task

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/command"
	"github.com/akijowski/aws-auto-alarm/internal/template"
)

const (
	eventbridgeEventSource     = "aws.tag"
	eventbridgeEventDetailType = "Tag Change on Resource"
)

type LambdaHook struct{}

func (h LambdaHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		
	}
}

type AlarmHandler struct {
	MetricAPI   autoalarm.MetricAlarmAPI
	ResourceAPI autoalarm.GetResourcesAPI
}

func (h *AlarmHandler) Handle(ctx context.Context, event *events.SQSEvent) (*events.SQSEventResponse, error) {
	logger := log.Ctx(ctx).With().
		Int("sqs_messages_count", len(event.Records)).
		Logger()
	logger.Info().Msg("Received SQS event")

	// do this for now, make better later
	for _, record := range event.Records {
		if err := h.handleSQSRecord(ctx, record); err != nil {
			logger.Error().Str("sqs_message_id", record.MessageId).Err(err).Msg("Failed to process SQS record")
			return nil, err
		}
	}

	return nil, nil
}

func (h *AlarmHandler) handleSQSRecord(ctx context.Context, record events.SQSMessage) error {
	logger := log.Ctx(ctx).With().
		Str("message_id", record.MessageId).
		Logger()
	logger.Info().Msg("Processing SQS record")

	event := new(events.EventBridgeEvent)
	if err := json.Unmarshal([]byte(record.Body), event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	logger = logger.With().Str("event_id", event.ID).Logger()
	logger.Debug().Interface("event", event).Msg("Unmarshalled event")

	logger.Info().
		Str("source", event.Source).
		Str("detail_type", event.DetailType).
		Strs("resources", event.Resources).
		Msg("Received EventBridge event")
	if err := filterEvent(event); err != nil {
		return fmt.Errorf("unable to process event: %w", err)
	}

	return buildAndRun(logger.WithContext(ctx), h.MetricAPI, event)
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

func buildAndRun(ctx context.Context, api autoalarm.MetricAlarmAPI, event *events.EventBridgeEvent) error {
	logger := log.Ctx(ctx)
	config, err := NewConfig(ctx, event)
	if err != nil {
		return fmt.Errorf("unable to create config: %w", err)
	}
	logger.Info().Interface("config", config).Msg("Created config")

	cmdRegistry := command.DefaultRegistry(api, logger)

	cmdType := "cloudwatch"
	if config.DryRun {
		cmdType = "json"
	}

	cmd, err := cmdRegistry.CreateCommand(ctx, cmdType, template.NewFileLoader(ctx, config))
	if config.Delete {
		cmd, err = cmdRegistry.DeleteCommand(ctx, cmdType, template.NewFileFinder(ctx, config))
	}
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	if err = cmd.Execute(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	logger.Info().Msg("event handling complete")
	return nil
}
