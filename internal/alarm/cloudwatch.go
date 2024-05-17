package alarm

import (
	"context"

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

func UpdateCloudwatch(ctx context.Context, client MetricAlarmAPI, delete bool, opts ...func(o *Options)) error {
	var err error

	opt := newOptions(opts...)

	alarms, err := generateAlarms(opt, embededTemplatesForARN(opt.ARN))
	if err != nil {
		return err
	}

	if delete {
		input := toAlarmDeletionInput(alarms)
		_, err = client.DeleteAlarms(ctx, input)
	} else {
		for _, alarm := range alarms {
			_, err = client.PutMetricAlarm(ctx, alarm)
		}
	}

	return err
}
