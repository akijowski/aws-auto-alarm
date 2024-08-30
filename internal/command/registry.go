package command

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	cmdcw "github.com/akijowski/aws-auto-alarm/internal/command/cloudwatch"
	"github.com/akijowski/aws-auto-alarm/internal/command/json"
)

type AlarmLoader interface {
	Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error)
}

type AlarmNameFinder interface {
	Find(ctx context.Context) ([]string, error)
}

// Registry is used to generate create and delete commands.
type Registry struct {
	api autoalarm.MetricAlarmAPI
	wr  io.Writer
}

// DefaultRegistry returns a new Registry using the provided api and writer.
func DefaultRegistry(api autoalarm.MetricAlarmAPI, wr io.Writer) *Registry {
	// use a map instead? map[string]Command embedded in the Registry
	return &Registry{
		api,
		wr,
	}
}

// CreateCommand returns an autoalarm.Command for creation or upsert based on the type t and the input from AlarmLoader.
func (r *Registry) CreateCommand(ctx context.Context, t string, l AlarmLoader) (autoalarm.Command, error) {
	in, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	switch t {
	case "json":
		return json.NewCreateCmd(in, r.wr), nil
	case "cloudwatch":
		return cmdcw.NewCreateCmd(in, r.api), nil
	default:
		return nil, fmt.Errorf("unsupported command type: %s", t)
	}
}

// DeleteCommand returns an autoalarm.Command for deletes based on the type t and the input from AlarmNameFinder.
func (r *Registry) DeleteCommand(ctx context.Context, t string, f AlarmNameFinder) (autoalarm.Command, error) {
	names, err := f.Find(ctx)
	if err != nil {
		return nil, err
	}

	in := &cloudwatch.DeleteAlarmsInput{AlarmNames: names}

	switch t {
	case "json":
		return json.NewDeleteCmd(in, r.wr), nil
	case "cloudwatch":
		return cmdcw.NewDeleteCmd(in, r.api), nil
	default:
		return nil, fmt.Errorf("unsupported command type: %s", t)
	}
}
