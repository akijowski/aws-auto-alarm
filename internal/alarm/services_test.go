package alarm

import (
	"context"
	"testing"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

func TestAddServiceData(t *testing.T) {
	t.Parallel()

	sqsARN, _ := awsarn.Parse(arnSQSValid)

	cases := map[string]struct {
		data   *Data
		wanted dataExtras
	}{
		"no service applies no data": {
			data: &Data{
				ARN:   awsarn.ARN{Service: "unknown"},
				Extra: make(dataExtras),
			},
			wanted: make(dataExtras),
		},
		"sqs applies data": {
			data: &Data{
				ARN:   sqsARN,
				Extra: make(dataExtras),
			},
			wanted: dataExtras{
				ExtraTagSQSQueueName: "queue",
				ExtraTagSQSDLQName:   "queue-dlq",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			assert := assert.New(t)

			err := AddServiceData(context.TODO(), tc.data)
			assert.NoError(err)

			assert.EqualValues(tc.wanted, tc.data.Extra)
		})
	}
}
