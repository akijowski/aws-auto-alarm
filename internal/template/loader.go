package template

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type ResourceMapper interface {
	Map(ctx context.Context) map[string]any
}

type FileLoader struct {
	config       *autoalarm.Config
	baseAlarm    *cloudwatch.PutMetricAlarmInput
	templateData *alarmData
}

func NewFileLoader(ctx context.Context, cfg *autoalarm.Config, rm ResourceMapper) *FileLoader {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("creating new file loader")
	data := newAlarmData(ctx, cfg, rm)
	log.Debug().Interface("alarm_data", data).Msg("alarm data created")
	return &FileLoader{
		config:       cfg,
		baseAlarm:    autoalarm.AlarmBase(cfg),
		templateData: data,
	}
}

// Load parses template.Template from the local file system using the configured autoalarm.Config, base Alarm, and alarmData.
func (f *FileLoader) Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("loading from file templates")
	tmpls, err := newTemplates(f.config.ParsedARN)
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %w", err)
	}

	log.Debug().Int("alarms_count", len(tmpls)).Msg("templates loaded")
	alarms := make([]*cloudwatch.PutMetricAlarmInput, 0)
	for _, tmpl := range tmpls {
		alarm, err := newAlarm(tmpl, f.templateData, f.baseAlarm)
		if err != nil {
			return nil, fmt.Errorf("unable to create alarm from template: %w", err)
		}
		alarms = append(alarms, alarm)
	}

	return alarms, nil
}
