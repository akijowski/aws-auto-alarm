package template

import (
	"context"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type FileFinder struct {
	config       *autoalarm.Config
	baseAlarm    *cloudwatch.PutMetricAlarmInput
	fs           fs.FS
	templateData *alarmData
}

func NewFileFinder(ctx context.Context, cfg *autoalarm.Config, rm ResourceMapper) *FileFinder {
	data := newAlarmData(ctx, cfg, rm)
	return &FileFinder{
		config:       cfg,
		baseAlarm:    autoalarm.AlarmBase(cfg),
		templateData: data,
		fs:           content,
	}
}

func (f *FileFinder) Find(_ context.Context) ([]string, error) {
	tmpls, err := templates(f.fs, f.config.ParsedARN)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, tmpl := range tmpls {
		alarm, err := newAlarm(tmpl, f.templateData, f.baseAlarm)
		if err != nil {
			return nil, err
		}
		names = append(names, aws.ToString(alarm.AlarmName))
	}

	return names, nil
}
