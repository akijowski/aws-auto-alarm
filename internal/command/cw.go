package command

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type PutMetricAlarmAPI interface {
	PutMetricAlarm(ctx context.Context, in *cloudwatch.PutMetricAlarmInput, opts ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricAlarmOutput, error)
}

type DeleteAlarmsAPI interface {
	DeleteAlarms(ctx context.Context, in *cloudwatch.DeleteAlarmsInput, opts ...func(*cloudwatch.Options)) (*cloudwatch.DeleteAlarmsOutput, error)
}

type MetricAlarmAPI interface {
	PutMetricAlarmAPI
	DeleteAlarmsAPI
}

type CWCmd struct {
	inputs   []*cloudwatch.PutMetricAlarmInput
	isDelete bool
	api      MetricAlarmAPI
}

func (b *Builder) NewCWCmd(api MetricAlarmAPI) *CWCmd {
	return &CWCmd{
		inputs:   b.inputs,
		isDelete: b.config.Delete,
		api:      api,
	}
}

func (c *CWCmd) Execute(ctx context.Context) error {
	if c.isDelete {
		alarmNames := make([]string, 0)
		for _, input := range c.inputs {
			alarmNames = append(alarmNames, *input.AlarmName)
		}

		deleteInput := &cloudwatch.DeleteAlarmsInput{AlarmNames: alarmNames}

		_, err := c.api.DeleteAlarms(ctx, deleteInput)
		if err != nil {
			return fmt.Errorf("unable to delete alarms: %w", err)
		}
	} else {
		for _, input := range c.inputs {
			_, err := c.api.PutMetricAlarm(ctx, input)
			if err != nil {
				return fmt.Errorf("unable to update alarm: %w", err)
			}
		}
	}

	return nil
}
