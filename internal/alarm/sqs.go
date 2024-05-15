package alarm

import (
	"fmt"
	"strings"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
)

func sqsResources(arn awsarn.ARN) (map[string]any, error) {
	queueName := strings.SplitN(arn.Resource, "/", 2)[1]

	return map[string]any{
		"DLQName":   fmt.Sprintf("%s-dlq", queueName),
		"QueueName": queueName,
	}, nil
}
