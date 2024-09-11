package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

func Test_sqsResources(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cfg    *config.Config
		given  map[string]any
		wanted map[string]any
	}{
		"does not modify map when service is not SQS": {
			cfg: &config.Config{
				ParsedARN: arn.ARN{
					Service: "dynamodb",
				},
			},
			given: map[string]any{
				"Foo": "Bar",
			},
			wanted: map[string]any{
				"Foo": "Bar",
			},
		},
		"adds queue info to map": {
			cfg: &config.Config{
				ParsedARN: arn.ARN{
					Service:  "sqs",
					Resource: "my-queue",
				},
			},
			given: map[string]any{},
			wanted: map[string]any{
				"QueueName": "my-queue",
				"DLQName":   "my-queue-dlq",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			sqsResources(tc.cfg, tc.given)

			assert.Equal(t, tc.wanted, tc.given)
		})
	}
}

func Test_queueNames(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		arn       arn.ARN
		overrides map[string]any
		wantQueue string
		wantDLQ   string
	}{
		"no override dlq is correct": {
			arn:       arn.ARN{Resource: "my-queue"},
			overrides: map[string]any{},
			wantQueue: "my-queue",
			wantDLQ:   "my-queue-dlq",
		},
		"override dlq is correct": {
			arn: arn.ARN{Resource: "other-queue"},
			overrides: map[string]any{
				"SQS_DLQ_NAME": "use-this-one",
			},
			wantQueue: "other-queue",
			wantDLQ:   "use-this-one",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			queue, dlq := queueNames(tc.arn, tc.overrides)

			assert := assert.New(t)

			assert.Equal(tc.wantQueue, queue)
			assert.Equal(tc.wantDLQ, dlq)
		})
	}
}
