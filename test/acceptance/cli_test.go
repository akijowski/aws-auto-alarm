package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akijowski/aws-auto-alarm/internal/cli"
	"github.com/akijowski/aws-auto-alarm/internal/config"
)

func TestOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		service   string
		config    func(testing.TB) (*config.Config, error)
		wantBytes func(testing.TB) ([]byte, error)
	}{
		{
			service:   "sqs",
			config:    configFromTestName,
			wantBytes: outputFromTestName,
		},
		{
			service:   "sqs_delete",
			config:    configFromTestName,
			wantBytes: outputFromTestName,
		},
		{
			service:   "sqs_override_dlq",
			config:    configFromTestName,
			wantBytes: outputFromTestName,
		},
		{
			service:   "tags",
			config:    configFromTestName,
			wantBytes: outputFromTestName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.service, func(t *testing.T) {
			t.Parallel()

			tc := tc

			require := require.New(t)

			config, err := tc.config(t)
			require.NoError(err)

			buf := new(bytes.Buffer)

			ctx := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t))).
				With().
				Caller().
				Str("arn", config.ParsedARN.String()).
				Logger().
				WithContext(context.Background())

			err = cli.New(config, nil, buf).Run(ctx)
			assert.NoError(t, err)

			b, err := tc.wantBytes(t)
			require.NoError(err)

			if config.Delete {
				wanted := new(cloudwatch.DeleteAlarmsInput)
				err = json.Unmarshal(b, wanted)
				require.NoError(err)

				actual := new(cloudwatch.DeleteAlarmsInput)
				err = json.Unmarshal(buf.Bytes(), actual)
				require.NoError(err)

				assert.ElementsMatch(t, wanted.AlarmNames, actual.AlarmNames)
			} else {
				wanted := make([]*cloudwatch.PutMetricAlarmInput, 0)
				err = json.Unmarshal(b, &wanted)
				require.NoError(err)

				actual := make([]*cloudwatch.PutMetricAlarmInput, 0)
				err = json.Unmarshal(buf.Bytes(), &actual)
				require.NoError(err)

				assert.ElementsMatch(t, wanted, actual)
			}
		})
	}

}

func configFromTestName(t testing.TB) (*config.Config, error) {
	t.Helper()

	fileName := fmt.Sprintf("./fixtures/input/%s.json", strings.SplitN(t.Name(), "/", 2)[1])
	t.Logf("filename: %s", fileName)

	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	config := new(config.Config)
	if err = json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	parsedARN, err := arn.Parse(config.ARN)
	if err != nil {
		return nil, err
	}
	config.ParsedARN = parsedARN

	return config, nil
}

func outputFromTestName(t testing.TB) ([]byte, error) {
	t.Helper()

	fileName := fmt.Sprintf("./fixtures/output/%s.json", strings.SplitN(t.Name(), "/", 2)[1])
	t.Logf("output filename: %s", fileName)

	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return file, nil
}
