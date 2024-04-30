package alarm

import (
	"context"
	"fmt"
)

const (
	ExtraTagSQSQueueName = "QueueName"
	ExtraTagSQSDLQName   = "DLQName"
)

func AddServiceData(ctx context.Context, d *Data) error {
	f := getMaps(d.ARN.Service)
	if f != nil {
		if err := f(ctx, d); err != nil {
			return err
		}
	}

	return nil
}

func getMaps(service string) DataOptionFunc {
	m := make(map[string]DataOptionFunc)
	m["sqs"] = sqsExtras

	return m[service]
}

func sqsExtras(_ context.Context, d *Data) error {
	queueName := d.ARN.Resource
	d.Extra[ExtraTagSQSQueueName] = queueName
	d.Extra[ExtraTagSQSDLQName] = fmt.Sprintf("%s-dlq", queueName)

	return nil
}
