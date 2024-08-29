package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type alarmData struct {
	AlarmPrefix string
	Resources   map[string]any
}

func newAlarmData(ctx context.Context, cfg *autoalarm.Config, m ResourceMapper) *alarmData {
	return &alarmData{
		AlarmPrefix: cfg.AlarmPrefix,
		Resources:   m.Map(ctx),
	}
}

func newAlarm(t *template.Template, data *alarmData, base *cloudwatch.PutMetricAlarmInput) (*cloudwatch.PutMetricAlarmInput, error) {
	buf := new(bytes.Buffer)

	input := new(cloudwatch.PutMetricAlarmInput)
	copyAlarmBase(base, input)

	if tags, ok := data.Resources["Tags"]; ok {
		input.Tags = append(input.Tags, tags.([]types.Tag)...)
	}

	if err := t.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("unable to template alarm: %w", err)
	}

	if err := json.Unmarshal(buf.Bytes(), input); err != nil {
		return nil, fmt.Errorf("unable to parse json: %w", err)
	}

	return input, nil
}

func copyAlarmBase(src, dest *cloudwatch.PutMetricAlarmInput) {
	dest.ActionsEnabled = src.ActionsEnabled
	dest.AlarmActions = src.AlarmActions
	dest.OKActions = src.OKActions
	dest.Tags = src.Tags
}
