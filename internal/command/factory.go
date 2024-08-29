package command

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	cwcmd "github.com/akijowski/aws-auto-alarm/internal/command/cloudwatch"
	"github.com/akijowski/aws-auto-alarm/internal/command/json"
	"github.com/akijowski/aws-auto-alarm/internal/resources"
	"github.com/akijowski/aws-auto-alarm/internal/template"
)

type AlarmLoader interface {
	Load(ctx context.Context) ([]*cloudwatch.PutMetricAlarmInput, error)
}

type AlarmNameFinder interface {
	Find(ctx context.Context) ([]string, error)
}

type NewCommandFn func(ctx context.Context, delete bool) (autoalarm.Command, error)

//TODO: Refactor all of this

// Factory is a factory for creating new commands.
type Factory struct {
	loader     AlarmLoader
	nameFinder AlarmNameFinder
	api        autoalarm.MetricAlarmAPI
	wr         io.Writer
}

func DefaultFactory(ctx context.Context, cfg *autoalarm.Config) *Factory {
	tmplLoader := template.NewFileLoader(ctx, cfg, resources.NewMapper(cfg))
	finder := template.NewFileFinder(ctx, cfg, resources.NewMapper(cfg))
	return &Factory{
		loader:     tmplLoader,
		nameFinder: finder,
	}
}

func (f *Factory) WithMetricAPI(api autoalarm.MetricAlarmAPI) NewCommandFn {
	f.api = api
	return func(ctx context.Context, delete bool) (autoalarm.Command, error) {
		if f.api == nil {
			return nil, fmt.Errorf("no API provided")
		}
		return f.newCommand(ctx, delete, "cloudwatch")
	}
}

func (f *Factory) WithWriter(wr io.Writer) NewCommandFn {
	f.wr = wr
	return func(ctx context.Context, delete bool) (autoalarm.Command, error) {
		if f.wr == nil {
			return nil, fmt.Errorf("no writer provided")
		}
		return f.newCommand(ctx, delete, "json")
	}
}

func (f *Factory) newCommand(ctx context.Context, delete bool, cmdType string) (autoalarm.Command, error) {
	if delete {
		return f.deleteCmd(ctx, cmdType)
	} else {
		return f.createCmd(ctx, cmdType)
	}
}

func (f *Factory) deleteCmd(ctx context.Context, cmdType string) (autoalarm.Command, error) {
	names, err := f.nameFinder.Find(ctx)
	if err != nil {
		return nil, err
	}
	input := &cloudwatch.DeleteAlarmsInput{AlarmNames: names}
	switch cmdType {
	case "json":
		return json.NewDeleteCmd(input, f.wr), nil
	case "cloudwatch":
		return cwcmd.NewDeleteCmd(input, f.api), nil
	default:
		return nil, fmt.Errorf("no matching command type %s", cmdType)
	}
}

func (f *Factory) createCmd(ctx context.Context, cmdType string) (autoalarm.Command, error) {
	inputs, err := f.loader.Load(ctx)
	if err != nil {
		return nil, err
	}
	switch cmdType {
	case "json":
		return json.NewCreateCmd(inputs, f.wr), nil
	case "cloudwatch":
		return cwcmd.NewCreateCmd(inputs, f.api), nil
	default:
		return nil, fmt.Errorf("no matching command type %s", cmdType)
	}
}
