package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type CreateCmd struct {
	inputs []*cloudwatch.PutMetricAlarmInput
	api    autoalarm.PutMetricAlarmAPI
}

func NewCreateCmd(inputs []*cloudwatch.PutMetricAlarmInput, api autoalarm.PutMetricAlarmAPI) *CreateCmd {
	return &CreateCmd{
		inputs: inputs,
		api:    api,
	}
}

func (c *CreateCmd) Execute(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("writing output to Cloudwatch")
	for _, in := range c.inputs {
		_, err := c.api.PutMetricAlarm(ctx, in)
		if err != nil {
			return err
		}
	}
	return nil
}

type DeleteCmd struct {
	input *cloudwatch.DeleteAlarmsInput
	api   autoalarm.DeleteAlarmsAPI
}

func NewDeleteCmd(input *cloudwatch.DeleteAlarmsInput, api autoalarm.DeleteAlarmsAPI) *DeleteCmd {
	return &DeleteCmd{
		input: input,
		api:   api,
	}
}

func (d *DeleteCmd) Execute(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("writing output to Cloudwatch")
	_, err := d.api.DeleteAlarms(ctx, d.input)
	if err != nil {
		return err
	}
	return nil
}
