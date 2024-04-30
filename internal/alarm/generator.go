package alarm

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

var (
	//go:embed templates/*
	content embed.FS
)

type CloudwatchAlarmAPI interface {
	CloudwatchPutMetricAlarmAPI
}

type CloudwatchPutMetricAlarmAPI interface {
	PutMetricAlarm(ctx context.Context, input *cloudwatch.PutMetricAlarmInput, opts ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricAlarmOutput, error)
}

type alarmGeneratorFn func(w io.Writer) error

type dataExtras map[string]any
type DataOptionFunc func(ctx context.Context, d *Data) error

type Data struct {
	ARN                 awsarn.ARN
	AlarmActions        []string
	InsufficientActions []string
	OKActions           []string
	Extra               dataExtras
}

func FromARN(ctx context.Context, arn string, opts ...DataOptionFunc) (*Data, error) {
	if !awsarn.IsARN(arn) {
		return nil, fmt.Errorf("provided arn %s is not valid", arn)
	}

	arnData, err := awsarn.Parse(arn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ARN %w", err)
	}

	data := &Data{ARN: arnData, Extra: make(dataExtras)}

	for _, opt := range opts {
		err = opt(ctx, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func WithData(d *Data) alarmGeneratorFn {
	return func(w io.Writer) error {
		t, err := template.ParseFS(content, fmt.Sprintf("templates/%s.json.tmpl", d.ARN.Service))
		if err != nil {
			return fmt.Errorf("failed to parse template for AWS service `%s`: %w", d.ARN.Service, err)
		}

		if err := t.Execute(w, d); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}
}

func WriteToCloudwatch(ctx context.Context, client CloudwatchAlarmAPI, g alarmGeneratorFn) error {
	_, err := marshalToInput(g)
	if err != nil {
		return err
	}

	return nil
}

func marshalToInput(g alarmGeneratorFn) ([]*cloudwatch.PutMetricAlarmInput, error) {
	buf := new(bytes.Buffer)

	if err := g(buf); err != nil {
		return nil, err
	}

	var input []*cloudwatch.PutMetricAlarmInput

	if err := json.NewDecoder(buf).Decode(&input); err != nil {
		return nil, err
	}

	return input, nil
}
