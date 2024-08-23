package template

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type ResourceMapper interface {
	Map() map[string]any
}

type FileLoader struct {
}

func (f *FileLoader) Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error) {
	// get templates
	// for each template, create inputs
	return nil, nil
}
