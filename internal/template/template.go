package template

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

var (
	//go:embed templates/*
	content embed.FS
)

type ResourceMapper interface {
	Map(ctx context.Context) map[string]any
}

// alarmData is what is applied to each alarm template.
type alarmData struct {
	// AlarmPrefix is an optional prefix for the alarm name.
	AlarmPrefix string
	// ARN is the ARN of the resource being templated.
	ARN arn.ARN
	// Resources is a map of data that can hold specific information based on the ARN resource type.
	Resources map[string]any
	// Tags is a map of tags to apply to the alarm.
	Tags map[string]string
}

func newAlarmData(ctx context.Context, cfg *config.Config, m ResourceMapper) *alarmData {
	return &alarmData{
		AlarmPrefix: cfg.AlarmPrefix,
		ARN:         cfg.ParsedARN,
		Resources:   m.Map(ctx),
		Tags:        cfg.Tags,
	}
}

func newAlarm(t *template.Template, data *alarmData, base *cloudwatch.PutMetricAlarmInput) (*cloudwatch.PutMetricAlarmInput, error) {
	buf := new(bytes.Buffer)

	input := new(cloudwatch.PutMetricAlarmInput)
	copyAlarmBase(base, input)

	applyTags(input, data)

	if err := t.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("unable to template alarm: %w", err)
	}

	if err := json.Unmarshal(buf.Bytes(), input); err != nil {
		return nil, fmt.Errorf("unable to parse json: %w", err)
	}

	return input, nil
}

func applyTags(input *cloudwatch.PutMetricAlarmInput, data *alarmData) {
	extraTags := []types.Tag{
		{
			Key:   aws.String("AWS_AUTO_ALARM_SOURCE_ARN"),
			Value: aws.String(data.ARN.String()),
		},
	}

	extraTags = append(extraTags, awsTags(data.Tags)...)

	input.Tags = append(input.Tags, extraTags...)
}

func awsTags(m map[string]string) []types.Tag {
	tags := make([]types.Tag, 0)
	for k, v := range m {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tags
}

func copyAlarmBase(src, dest *cloudwatch.PutMetricAlarmInput) {
	dest.ActionsEnabled = src.ActionsEnabled
	dest.AlarmActions = src.AlarmActions
	dest.OKActions = src.OKActions
	dest.Tags = src.Tags
}
