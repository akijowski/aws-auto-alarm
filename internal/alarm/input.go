package alarm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func alarmBaseFromOptions(opt *Options) *cloudwatch.PutMetricAlarmInput {
	return &cloudwatch.PutMetricAlarmInput{
		ActionsEnabled: aws.Bool(true),
		OKActions:      opt.OKActions,
		AlarmActions:   opt.AlarmActions,
		Tags: []types.Tag{
			{
				Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
				Value: aws.String("true"),
			},
		},
	}
}

func generateAlarms(arn awsarn.ARN, opt *Options) ([]*cloudwatch.PutMetricAlarmInput, error) {
	alarmInputs := make([]*cloudwatch.PutMetricAlarmInput, 0)

	d, err := newAlarmData(arn, opt)
	if err != nil {
		return nil, err
	}

	tmpls, err := templatesForARN(arn)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range tmpls {
		alarm, err := templateAlarm(tmpl, opt, d)
		if err != nil {
			return nil, err
		}

		alarmInputs = append(alarmInputs, alarm)
	}

	return alarmInputs, nil
}

func toAlarmDeletionInput(alarms []*cloudwatch.PutMetricAlarmInput) *cloudwatch.DeleteAlarmsInput {
	deletes := make([]string, 0)

	for _, alarm := range alarms {
		deletes = append(deletes, aws.ToString(alarm.AlarmName))
	}

	return &cloudwatch.DeleteAlarmsInput{AlarmNames: deletes}
}
