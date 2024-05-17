package alarm

import (
	"errors"
	"testing"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestAlarmBaseFromOptions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		given *Options
		want  *cloudwatch.PutMetricAlarmInput
	}{
		"default is returned with no opts": {
			want: &cloudwatch.PutMetricAlarmInput{
				ActionsEnabled: aws.Bool(true),
				Tags: []types.Tag{
					{
						Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
						Value: aws.String("true"),
					},
				},
			},
		},
		"ok actions are added": {
			given: &Options{OKActions: []string{"foo"}},
			want: &cloudwatch.PutMetricAlarmInput{
				ActionsEnabled: aws.Bool(true),
				OKActions:      []string{"foo"},
				Tags: []types.Tag{
					{
						Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
						Value: aws.String("true"),
					},
				},
			},
		},
		"alarm actions are added": {
			given: &Options{AlarmActions: []string{"foo"}},
			want: &cloudwatch.PutMetricAlarmInput{
				ActionsEnabled: aws.Bool(true),
				AlarmActions:   []string{"foo"},
				Tags: []types.Tag{
					{
						Key:   aws.String("AWS_AUTO_ALARM_MANAGED"),
						Value: aws.String("true"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			actual := alarmBaseFromOptions(tc.given)

			assert.Equal(t, tc.want, actual)
		})
	}
}

func TestGenerateAlarms(t *testing.T) {

	validOpts := &Options{ARN: arn.ARN{
		Service:  "sqs",
		Resource: "queue/queue-1",
	}}

	emptyGenerator := func() ([]*template.Template, error) {
		return make([]*template.Template, 0), nil
	}

	validTmpl := template.Must(template.New("valid").Parse("{}"))

	brokenTmpl := template.Must(template.New("broken").Parse("invalid"))

	cases := map[string]struct {
		givenOpts    *Options
		givenGenFunc func(*testing.T) templateGeneratorFunc
		wantErr      bool
		want         []*cloudwatch.PutMetricAlarmInput
	}{
		"unsupported ARN returns error": {
			givenOpts: &Options{ARN: arn.ARN{
				Service: "dne",
			}},
			givenGenFunc: func(t *testing.T) templateGeneratorFunc {
				return emptyGenerator
			},
			wantErr: true,
		},
		"template generation error returns error": {
			givenOpts: validOpts,
			givenGenFunc: func(t *testing.T) templateGeneratorFunc {
				t.Helper()

				return func() ([]*template.Template, error) {
					return nil, errors.New("failed")
				}
			},
			wantErr: true,
		},
		"template error returns error": {
			givenOpts: validOpts,
			givenGenFunc: func(t *testing.T) templateGeneratorFunc {
				return func() ([]*template.Template, error) {
					return []*template.Template{brokenTmpl}, nil
				}
			},
			wantErr: true,
		},
		"valid template is added": {
			givenOpts: validOpts,
			givenGenFunc: func(t *testing.T) templateGeneratorFunc {
				return func() ([]*template.Template, error) {
					return []*template.Template{validTmpl}, nil
				}
			},
			want: []*cloudwatch.PutMetricAlarmInput{
				alarmBaseFromOptions(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			actual, err := generateAlarms(tc.givenOpts, tc.givenGenFunc(t))
			t.Log(err)

			assert := assert.New(t)

			assert.Equal(tc.wantErr, err != nil)

			if err == nil {
				t.Log(spew.Sdump(actual))
				assert.Equal(tc.want, actual)
			}
		})
	}
}

func TestToAlarmDeletionInput(t *testing.T) {
	given := []*cloudwatch.PutMetricAlarmInput{
		{AlarmName: aws.String("alarm-1")},
		{AlarmName: aws.String("alarm-2")},
	}

	want := &cloudwatch.DeleteAlarmsInput{
		AlarmNames: []string{
			"alarm-1",
			"alarm-2",
		},
	}

	actual := toAlarmDeletionInput(given)

	assert.Equal(t, want, actual)
}
