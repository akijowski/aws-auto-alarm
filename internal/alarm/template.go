package alarm

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

var (
	//go:embed templates/*
	content embed.FS
)

// alarmData is the struct passed to the embeded templates
type alarmData struct {
	AlarmPrefix string
	Resources   map[string]any
}

func newAlarmData(arn awsarn.ARN, opt *Options) (*alarmData, error) {
	d := &alarmData{
		AlarmPrefix: opt.AlarmPrefix,
	}

	switch arn.Service {
	case "sqs":
		resources, err := sqsResources(arn)
		if err != nil {
			return nil, err
		}
		d.Resources = resources
	default:
		return nil, fmt.Errorf("unable to create resource data for service: %s", arn.Service)
	}

	return d, nil
}

func templatesForARN(arn awsarn.ARN) ([]*template.Template, error) {
	t, err := template.ParseFS(content, fmt.Sprintf("templates/%s/*", arn.Service))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	return t.Templates(), nil
}

func templateAlarm(tmpl *template.Template, opt *Options, d *alarmData) (*cloudwatch.PutMetricAlarmInput, error) {
	alarm := alarmBaseFromOptions(opt)

	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, d)
	if err != nil {
		return nil, fmt.Errorf("template error: %w", err)
	}

	err = json.Unmarshal(buf.Bytes(), &alarm)
	if err != nil {
		return nil, fmt.Errorf("json error: %w", err)
	}

	return alarm, nil
}
