package autoalarm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
)

// Command is a generic interface that can be used by cli or lambda environments.
type Command interface {
	Execute(context.Context) error
}
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
type GetResourcesAPI interface {
	GetResources(ctx context.Context, in *resourcegroupstaggingapi.GetResourcesInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}
