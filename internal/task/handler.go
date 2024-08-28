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
	"github.com/akijowski/aws-auto-alarm/internal/awsclient"
	"github.com/akijowski/aws-auto-alarm/internal/command"
)

const (
	eventbridgeEventSource     = "aws.tag"
	eventbridgeEventDetailType = "Tag Change on Resource"
)

type AlarmHandler struct {
}

func (h *AlarmHandler) Handle(ctx context.Context, event *events.SQSEvent) (*events.SQSEventResponse, error) {
	log := zerolog.Ctx(ctx).With().
		Int("sqs_messages_count", len(event.Records)).
		Logger()
	log.Info().Msg("Received SQS event")

	cw, err := awsclient.CloudWatch(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create CloudWatch client")
		return nil, err
	}

	// do this for now, make better later
	for _, record := range event.Records {
		if err := handleSQSRecord(ctx, cw, record); err != nil {
			log.Error().Str("sqs_message_id", record.MessageId).Err(err).Msg("Failed to process SQS record")
			return nil, err
		}
	}

	return nil, nil
}

func handleSQSRecord(ctx context.Context, api command.MetricAlarmAPI, record events.SQSMessage) error {
	log := zerolog.Ctx(ctx).With().
		Str("message_id", record.MessageId).
		Logger()
	log.Info().Msg("Processing SQS record")

	event := new(events.EventBridgeEvent)
	if err := json.Unmarshal([]byte(record.Body), event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	log.Info().Str("source", event.Source).Str("detail_type", event.DetailType).Msg("Received EventBridge event")
	if err := filterEvent(event); err != nil {
		return fmt.Errorf("unable to process event: %w", err)
	}

	return buildAndRun(ctx, api, log, event)
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

func buildAndRun(ctx context.Context, api command.MetricAlarmAPI, wr io.Writer, event *events.EventBridgeEvent) error {
	config, err := autoalarm.NewLambdaConfig(ctx, event)
	if err != nil {
		return fmt.Errorf("unable to create config: %w", err)
	}

	builder, err := command.DefaultBuilder(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to create command builder: %w", err)
	}

	var cmd autoalarm.Command
	if config.DryRun {
		cmd = builder.NewJSONCmd(wr)
	} else {
		cmd = builder.NewCWCmd(api)
	}

	if err = cmd.Execute(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	return nil
}
