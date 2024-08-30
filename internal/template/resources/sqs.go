package resources

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

func sqsResources(cfg *autoalarm.Config, m map[string]any) {
	arn := cfg.ParsedARN
	if arn.Service == "sqs" {
		queue, dlq := queueNames(arn, cfg.Overrides)
		m["QueueName"] = queue
		m["DLQName"] = dlq
	}
}

func queueNames(a arn.ARN, overrides map[string]any) (string, string) {
	queue := a.Resource
	dlq := fmt.Sprintf("%s-dlq", queue)

	if dlqOverride, ok := overrides["SQS_DLQ_NAME"]; ok {
		dlq = dlqOverride.(string)
	}

	return queue, dlq

}
