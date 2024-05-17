package alarm

import (
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type templateGeneratorFunc func() ([]*template.Template, error)

func alarmBaseFromOptions(opt *Options) *cloudwatch.PutMetricAlarmInput {
	base := &cloudwatch.PutMetricAlarmInput{
		ActionsEnabled: aws.Bool(true),
		Tags: []types.Tag{
			{
				Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
				Value: aws.String("true"),
			},
		},
	}

	if opt != nil {
		base.OKActions = opt.OKActions
		base.AlarmActions = opt.AlarmActions
	}

	return base
}

func generateAlarms(opt *Options, tgen templateGeneratorFunc) ([]*cloudwatch.PutMetricAlarmInput, error) {
	alarmInputs := make([]*cloudwatch.PutMetricAlarmInput, 0)

	d, err := newAlarmData(opt)
	if err != nil {
		return nil, err
	}

	tmpls, err := tgen()
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
