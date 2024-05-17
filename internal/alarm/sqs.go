package alarm

import (
	"fmt"
	"strings"
)

func sqsResources(opt *Options) (map[string]any, error) {
	resources := make(map[string]any)

	queueName := strings.SplitN(opt.ARN.Resource, "/", 2)[1]
	dlqName := fmt.Sprintf("%s-dlq", queueName)

	if dlqNameOverride, ok := opt.Overrides["SQS_DLQ_NAME"]; ok {
		dlqName = dlqNameOverride.(string)
	}

	resources["DLQName"] = dlqName
	resources["QueueName"] = queueName

	return resources, nil
}
