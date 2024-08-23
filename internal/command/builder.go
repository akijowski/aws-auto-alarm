package command

import (
	"context"
	"fmt"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/resources"
	"github.com/akijowski/aws-auto-alarm/internal/template"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type AlarmLoader interface {
	Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error)
}

type Builder struct {
	config *autoalarm.Config
	inputs []*cloudwatch.PutMetricAlarmInput
}

func DefaultBuilder(ctx context.Context, cfg *autoalarm.Config) (*Builder, error) {
	b := &Builder{
		config: cfg,
	}

	tmplLoader := template.NewFileLoader(ctx, cfg, resources.NewMapper(cfg))
	inputs, err := tmplLoader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load alarm input: %w", err)
	}
	b.inputs = inputs

	return b, nil
}
