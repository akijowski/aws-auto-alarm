package autoalarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLambdaConfig(t *testing.T) {
	t.Parallel()

	ctx := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t))).
		With().Caller().Logger().WithContext(context.Background())

	detail := &tagChangeDetail{
		Tags: map[string]string{
			"AWS_AUTO_ALARM_ALARMPREFIX": "test",
		},
	}
	detailBytes, err := json.Marshal(detail)
	require.NoError(t, err)

	event := &events.EventBridgeEvent{ID: "event12345"}

	t.Run("no resources in event returns error", func(t *testing.T) {

		_, err = NewLambdaConfig(ctx, event)
		assert.Error(t, err)
	})

	t.Run("unable to parse ARN returns error", func(t *testing.T) {

		event.Resources = []string{"invalid-arn"}
		_, err = NewLambdaConfig(ctx, event)
		assert.Error(t, err)
	})

	t.Run("unable to unmarshal detail returns error", func(t *testing.T) {

		event.Resources = []string{"arn:aws:sqs:us-east-1:123456789012:test-queue"}
		event.Detail = []byte("invalid-detail")
		_, err = NewLambdaConfig(ctx, event)
		assert.Error(t, err)
	})

	t.Run("returns correct config", func(t *testing.T) {

		event.Resources = []string{"arn:aws:sqs:us-east-1:123456789012:test-queue"}
		event.Detail = detailBytes
		cfg, err := NewLambdaConfig(ctx, event)
		assert.NoError(t, err)

		wanted := &Config{
			AlarmPrefix: "test",
			ARN:         "arn:aws:sqs:us-east-1:123456789012:test-queue",
			ParsedARN: arn.ARN{
				Partition: "aws",
				Service:   "sqs",
				Region:    "us-east-1",
				AccountID: "123456789012",
				Resource:  "test-queue",
			},
		}
		assert.Equal(t, wanted, cfg)
	})
}

func Test_parseDetail(t *testing.T) {
	t.Parallel()

	defaultQueueARN := arn.ARN{
		Partition: "aws",
		Service:   "sqs",
		Region:    "us-east-1",
		AccountID: "123456789012",
		Resource:  "test-queue",
	}

	cases := map[string]struct {
		given   func(t testing.TB) *tagChangeDetail
		cfg     func(t testing.TB) *Config
		want    *Config
		wantErr bool
	}{
		"basic info is configured": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_ALARMPREFIX": "test",
						"AWS_AUTO_ALARM_DRYRUN":      "true",
						"AWS_AUTO_ALARM_QUIET":       "true",
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			want: &Config{
				AlarmPrefix: "test",
				DryRun:      true,
				Quiet:       true,
				ParsedARN:   defaultQueueARN,
			},
		},
		"delete is configured": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					ChangedTagKeys: []string{
						"AWS_AUTO_ALARM_ENABLED",
					},
					Tags: map[string]string{
						"FOO": "BAR",
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			want: &Config{
				Delete:    true,
				ParsedARN: defaultQueueARN,
			},
		},
		"actions are configured": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_OKACTIONS":    "sns1,sns2",
						"AWS_AUTO_ALARM_ALARMACTIONS": "sns3",
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			want: &Config{
				Delete:       false,
				ParsedARN:    defaultQueueARN,
				OKActions:    []string{"sns1", "sns2"},
				AlarmActions: []string{"sns3"},
			},
		},
		"overrides are configured": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_OVERRIDES": `{"SQS_DLQ_NAME":"test-queue-dlq","other":"value"}`,
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			want: &Config{
				ParsedARN: defaultQueueARN,
				Overrides: map[string]any{
					"SQS_DLQ_NAME": "test-queue-dlq",
					"other":        "value",
				},
			},
		},
		"tags are configured": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_TAGS": `{"tag1":"value1","tag2":"value2"}`,
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			want: &Config{
				ParsedARN: defaultQueueARN,
				Tags: map[string]string{
					"tag1": "value1",
					"tag2": "value2",
				},
			},
		},
		"invalid overrides returns error": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_OVERRIDES": `{"SQS_DLQ_NAME":"test-queue-dlq","other":"value"`,
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			wantErr: true,
		},
		"invalid tags returns error": {
			given: func(t testing.TB) *tagChangeDetail {
				return &tagChangeDetail{
					Tags: map[string]string{
						"AWS_AUTO_ALARM_TAGS": `["tag1":"value1"]`,
					},
				}
			},
			cfg: func(t testing.TB) *Config {
				return &Config{
					ParsedARN: defaultQueueARN,
				}
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			detail := tc.given(t)
			cfg := tc.cfg(t)

			ctx := zerolog.New(
				zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t)),
			).With().Caller().Logger().WithContext(context.Background())

			err := parseDetail(ctx, cfg, detail)

			assert.Equal(t, tc.wantErr, err != nil)

			if err != nil {
				t.Logf("error: %v", err)
			}

			if !tc.wantErr {
				assert.Equal(t, tc.want, cfg)
			}
		})
	}
}
