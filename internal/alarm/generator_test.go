package alarm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	arnSQSValid = "arn:aws:sqs:us-east-2:444455556666:queue"
)

func TestFromARN(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		arn     string
		wantErr bool
	}{
		"invalid ARN returns error": {
			arn:     "test:arn",
			wantErr: true,
		},
		"valid ARN returns data": {
			arn: arnSQSValid,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			assert := assert.New(t)

			actual, err := FromARN(context.TODO(), tc.arn)

			if tc.wantErr {
				assert.Error(err)
				t.Log(err)
			} else {
				assert.NoError(err)
				assert.NotNil(actual.ARN.Resource)
				assert.NotNil(actual.Extra)
			}
		})
	}
}

func TestFromARN_options(t *testing.T) {
	t.Parallel()

	sqsARN, _ := awsarn.Parse(arnSQSValid)

	cases := map[string]struct {
		givenFunc DataOptionFunc
		wantData  *Data
		wantErr   bool
	}{
		"option func is applied successfully": {
			givenFunc: func(ctx context.Context, d *Data) error {
				d.Extra["test"] = "success"

				return nil
			},
			wantData: &Data{
				ARN:   sqsARN,
				Extra: dataExtras{"test": "success"},
			},
		},
		"option func error is returned": {
			givenFunc: func(ctx context.Context, d *Data) error {
				return errors.New("an error occurred")
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			actual, err := FromARN(context.TODO(), arnSQSValid, tc.givenFunc)

			assert := assert.New(t)

			if tc.wantErr {
				assert.Error(err)
				t.Log(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantData, actual)
			}
		})
	}
}

func TestWithData(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		data    *Data
		wantErr bool
	}{
		"invalid aws service returns error": {
			data: &Data{
				ARN: awsarn.ARN{Service: "unknown"},
			},
			wantErr: true,
		},
		"template error returns error": {
			data: &Data{
				ARN: awsarn.ARN{Service: "test-invalid"},
			},
			wantErr: true,
		},
		"valid template returns no error": {
			data: &Data{
				ARN: awsarn.ARN{Service: "test-valid"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			assert := assert.New(t)

			buf := new(bytes.Buffer)

			err := WithData(tc.data)(buf)

			if tc.wantErr {
				assert.Error(err)
				t.Log(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestWithData_services(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		arn    func() awsarn.ARN
		extras dataExtras
	}{
		"sqs is valid": {
			arn: func() awsarn.ARN {
				a, _ := awsarn.Parse(arnSQSValid)
				return a
			},
			extras: dataExtras{
				ExtraTagSQSQueueName: "queue-a",
				ExtraTagSQSDLQName:   "queue-a-dlq",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			assert := assert.New(t)

			d := &Data{
				ARN:   tc.arn(),
				Extra: tc.extras,
			}

			buf := new(bytes.Buffer)

			err := WithData(d)(buf)

			assert.NoError(err, "failed to write")

			b, err := io.ReadAll(buf)
			require.NoError(t, err)

			t.Logf("%s\n", b)

			var input []*cloudwatch.PutMetricAlarmInput

			err = json.Unmarshal(b, &input)

			assert.NoError(err, "failed to marshal to cloudwatch input")
		})
	}
}
