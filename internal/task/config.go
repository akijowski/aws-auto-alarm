package task

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog/log"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

type tagChangeDetail struct {
	ChangedTagKeys []string          `json:"changed-tag-keys"`
	Service        string            `json:"service"`
	ResourceType   string            `json:"resource-type"`
	Version        float64           `json:"version"`
	Tags           map[string]string `json:"tags"`
}

// NewConfig creates a new Config from an EventBridge event.
// The event must have a single resource ARN and a detail field that
// contains the tag change details.
func NewConfig(ctx context.Context, event *events.EventBridgeEvent) (*config.Config, error) {
	logger := log.Ctx(ctx)

	logger.Debug().Msg("parsing cfg from event")
	cfg := new(config.Config)

	if len(event.Resources) == 0 {
		return nil, fmt.Errorf("no resources in event")
	}

	parsedARN, err := arn.Parse(event.Resources[0])
	if err != nil {
		return nil, fmt.Errorf("unable to parse ARN: %w", err)
	}
	cfg.ARN = event.Resources[0]
	cfg.ParsedARN = parsedARN

	detail := new(tagChangeDetail)
	if err := json.Unmarshal(event.Detail, detail); err != nil {
		return nil, fmt.Errorf("unable to unmarshal detail: %w", err)
	}

	if err = parseDetail(ctx, cfg, detail); err != nil {
		return nil, fmt.Errorf("unable to parse detail: %w", err)
	}

	return cfg, nil
}

func parseDetail(ctx context.Context, cfg *config.Config, detail *tagChangeDetail) error {
	logger := log.Ctx(ctx)
	logger.Debug().Interface("detail", detail).Msg("processing tag change")

	cfg.Delete = isDeleteAction(detail)

	parseDetailTags(detail.Tags, cfg)

	alarmActions(detail, cfg)

	if overrides, ok := detail.Tags["AWS_AUTO_ALARM_OVERRIDES"]; ok {
		if err := json.Unmarshal([]byte(overrides), &cfg.Overrides); err != nil {
			return err
		}
	}

	if tags, ok := detail.Tags["AWS_AUTO_ALARM_TAGS"]; ok {
		if err := json.Unmarshal([]byte(tags), &cfg.Tags); err != nil {
			return err
		}
	}

	logger.Debug().Interface("config", cfg).Msg("configuration complete")
	return nil
}

func parseDetailTags(detailTags map[string]string, cfg *config.Config) {
	// This is ugly but we can fix it later
	for key, value := range detailTags {
		switch key {
		case "AWS_AUTO_ALARM_ALARMPREFIX":
			cfg.AlarmPrefix = value
		case "AWS_AUTO_ALARM_DRYRUN":
			cfg.DryRun = value == "true"
		case "AWS_AUTO_ALARM_QUIET":
			cfg.Quiet = value == "true"
		}
	}
}

func isDeleteAction(detail *tagChangeDetail) bool {
	enabledIsChanged := slices.Contains(detail.ChangedTagKeys, "AWS_AUTO_ALARM_ENABLED")
	_, enabledIsPresent := detail.Tags["AWS_AUTO_ALARM_ENABLED"]

	return enabledIsChanged && !enabledIsPresent
}

func alarmActions(detail *tagChangeDetail, cfg *config.Config) {
	if alarmActionsStr, ok := detail.Tags["AWS_AUTO_ALARM_ALARMACTIONS"]; ok {
		cfg.AlarmActions = strings.Split(alarmActionsStr, ",")
	}
	if okActionsStr, ok := detail.Tags["AWS_AUTO_ALARM_OKACTIONS"]; ok {
		cfg.OKActions = strings.Split(okActionsStr, ",")
	}
}
