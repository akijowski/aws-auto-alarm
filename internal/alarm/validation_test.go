package alarm

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		arn arn.ARN
	}{
		"valid eventbridge passes": {
			arn: arn.ARN{
				Service:  "events",
				Resource: "rule/my-cool-bus/this-rules",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			err := IsValid(tc.arn)

			assert.NoError(t, err)
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		validator func(arn.ARN) error
		wantErr   bool
	}{
		"validation success returns no error": {
			validator: func(a arn.ARN) error {
				return nil
			},
		},
		"validation failure returns error": {
			validator: func(a arn.ARN) error {
				return errors.New("nope")
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			err := validate(arn.ARN{}, tc.validator)

			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestEventBridgeValidator(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		arn     arn.ARN
		wantErr bool
	}{
		"not eventbridge returns no error": {
			arn: arn.ARN{Service: "sqs"},
		},
		"eventbridge rule returns no error": {
			arn: arn.ARN{
				Service:  "events",
				Resource: "rule/my-cool-bus/this-rules",
			},
		},
		"eventbridge rule alt returns no error": {
			arn: arn.ARN{
				Service:  "events",
				Resource: "rule/this-rules",
			},
		},
		"eventbridge unsupported type returns error": {
			arn: arn.ARN{
				Service:  "events",
				Resource: "event-bus/my-cool-bus",
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			err := eventBridgeValidator()(tc.arn)

			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}
