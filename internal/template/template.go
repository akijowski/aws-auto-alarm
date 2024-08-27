package template

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

var (
	//go:embed templates/*
	content embed.FS
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

func newTemplates(arn awsarn.ARN) ([]*template.Template, error) {
	t, err := template.ParseFS(content, fmt.Sprintf("templates/%s/*", arn.Service))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	return t.Templates(), nil
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
