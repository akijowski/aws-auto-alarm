package autoalarm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

func Test_parseARN(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		givenARN string
		wantARN  arn.ARN
		wantErr  bool
	}{
		"empty ARN returns error": {
			givenARN: "",
			wantErr:  true,
		},
		"invalid ARN returns error": {
			givenARN: "invalid",
			wantErr:  true,
		},
		"valid ARN returns ARN": {
			givenARN: "arn:aws:cloudwatch:us-west-2:123456789012:alarm/my-alarm",
			wantARN: arn.ARN{
				Partition: "aws",
				Service:   "cloudwatch",
				Region:    "us-west-2",
				AccountID: "123456789012",
				Resource:  "alarm/my-alarm",
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				ARN: tc.givenARN,
			}

			err := parseARN(cfg)

			assert.Equal(t, tc.wantErr, err != nil)

			if !tc.wantErr {
				assert.Equal(t, tc.wantARN, cfg.ParsedARN)
			}
		})
	}
}
