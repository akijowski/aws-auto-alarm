package command

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

func deleteInput(alarms []*cloudwatch.PutMetricAlarmInput) *cloudwatch.DeleteAlarmsInput {
	alarmNames := make([]string, 0)
	for _, input := range alarms {
		name := aws.ToString(input.AlarmName)
		alarmNames = append(alarmNames, name)
	}

	return &cloudwatch.DeleteAlarmsInput{AlarmNames: alarmNames}
}
