package template

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

var (
	//go:embed templates/*
	content embed.FS
)

type alarmData struct {
	AlarmPrefix string
	Resources   map[string]any
}

type Parser struct {
	baseAlarm *cloudwatch.PutMetricAlarmInput
	data      *alarmData
}

func NewParser(cfg *autoalarm.Config) *Parser {
	return &Parser{
		baseAlarm: autoalarm.AlarmBase(cfg),
		data:      &alarmData{},
	}
}

func newTemplates(arn awsarn.ARN) ([]*template.Template, error) {
	t, err := template.ParseFS(content, fmt.Sprintf("templates/%s/*", arn.Service))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	return t.Templates(), nil
}

func (p *Parser) newAlarm(t *template.Template) (*cloudwatch.PutMetricAlarmInput, error) {
	buf := new(bytes.Buffer)

	input := new(cloudwatch.PutMetricAlarmInput)
	copyAlarmBase(p.baseAlarm, input)

	if err := t.Execute(buf, p.data); err != nil {
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
