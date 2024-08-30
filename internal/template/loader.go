package template

import (
	"context"
	"fmt"
	"io/fs"
	"text/template"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/template/resources"
)

// FileLoader loads alarm input based on the embedded fs.FS of templates and a base alarm generated with the provided
// autoalarm.Config.
type FileLoader struct {
	config       *autoalarm.Config
	baseAlarm    *cloudwatch.PutMetricAlarmInput
	fs           fs.FS
	templateData *alarmData
}

func NewFileLoader(ctx context.Context, cfg *autoalarm.Config) *FileLoader {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("creating new file loader")
	data := newAlarmData(ctx, cfg, resources.NewMapper(cfg))
	log.Debug().Interface("alarm_data", data).Msg("alarm data created")
	return &FileLoader{
		config:       cfg,
		baseAlarm:    autoalarm.AlarmBase(cfg),
		templateData: data,
		fs:           content,
	}
}

func templates(content fs.FS, arn awsarn.ARN) ([]*template.Template, error) {
	tmpls, err := template.ParseFS(content, fmt.Sprintf("templates/%s/*", arn.Service))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	return tmpls.Templates(), nil
}

// Load parses template.Template from the local file system using the configured autoalarm.Config, base Alarm, and alarmData.
func (f *FileLoader) Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("loading from file templates")
	tmpls, err := templates(f.fs, f.config.ParsedARN)
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
