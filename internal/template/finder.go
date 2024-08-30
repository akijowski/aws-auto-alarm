package template

import (
	"context"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/template/resources"
)

// FileFinder returns a slice of alarm names by applying the autoalarm.Config to the provided templates in the
// embedded fs.FS.
type FileFinder struct {
	config       *autoalarm.Config
	baseAlarm    *cloudwatch.PutMetricAlarmInput
	fs           fs.FS
	templateData *alarmData
}

func NewFileFinder(ctx context.Context, cfg *autoalarm.Config) *FileFinder {
	data := newAlarmData(ctx, cfg, resources.NewMapper(cfg))
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
