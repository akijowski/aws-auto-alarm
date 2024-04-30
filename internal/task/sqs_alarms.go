package task

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type CloudWatchPutMetricAlarmAPI interface {
	PutMetricAlarm(ctx context.Context, input *cloudwatch.PutMetricAlarmInput, opts ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricAlarmOutput, error)
}

type CloudWatchDeleteAlarmsAPI interface {
	DeleteMetricAlarm(ctx context.Context, input *cloudwatch.DeleteAlarmsInput, opts ...func(*cloudwatch.Options)) (*cloudwatch.DeleteAlarmsOutput, error)
}

type CloudWatchAlarmAPI interface {
	CloudWatchPutMetricAlarmAPI
	CloudWatchDeleteAlarmsAPI
}
