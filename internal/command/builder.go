package command

import (
	"context"
	"fmt"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type AlarmLoader interface {
	Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error)
}

type Builder struct {
	config *autoalarm.Config
	inputs []*cloudwatch.PutMetricAlarmInput
}

func Build(ctx context.Context, cfg *autoalarm.Config, l AlarmLoader) (*Builder, error) {
	b := new(Builder)
	b.config = cfg

	inputs, err := l.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load alarm input: %w", err)
	}
	b.inputs = inputs

	return b, nil
}
