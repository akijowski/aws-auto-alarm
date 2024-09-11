package template

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

// alarmBase returns a cloudwatch.PutMetricAlarmInput that will be applied to all generated alarms.
func alarmBase(cfg *config.Config) *cloudwatch.PutMetricAlarmInput {
	base := &cloudwatch.PutMetricAlarmInput{
		ActionsEnabled: aws.Bool(true),
		Tags: []types.Tag{
			{
				Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
				Value: aws.String("true"),
			},
		},
	}

	if cfg != nil {
		base.OKActions = cfg.OKActions
		base.AlarmActions = cfg.AlarmActions
	}

	return base
}
