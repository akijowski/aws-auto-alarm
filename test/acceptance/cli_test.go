package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type cliTestCase struct {
	Config *config.Config  `json:"input"`
	Output json.RawMessage `json:"output"`
}

func TestOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
	}{
		{
			name:     "sqs",
			fileName: "fixtures/cli/sqs.json",
		},
		{
			name:     "sqs_delete",
			fileName: "fixtures/cli/sqs_delete.json",
		},
		{
			name:     "sqs_override_dlq",
			fileName: "fixtures/cli/sqs_override_dlq.json",
		},
		{
			name:     "tags",
			fileName: "fixtures/cli/tags.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc := tc

			require := require.New(t)

			testCase, err := loadTestCase(t, tc.fileName)
			require.NoError(err)

			config := testCase.Config
			require.NotNil(config)

			b := testCase.Output

			buf := new(bytes.Buffer)

			ctx := zerolog.New(zerolog.NewTestWriter(t)).
				With().
				Caller().
				Timestamp().
				Logger().
				WithContext(context.Background())

			c := cli.New(config, nil, buf)
			err = c.Run(ctx)
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

func loadTestCase(t testing.TB, fileName string) (*cliTestCase, error) {
	t.Helper()
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	tc := new(cliTestCase)
	if err = json.Unmarshal(b, tc); err != nil {
		return nil, err
	}

	if tc.Config.ARN != "" {
		parsedARN, err := arn.Parse(tc.Config.ARN)
		if err != nil {
			return nil, err
		}
		tc.Config.ParsedARN = parsedARN
	}

	return tc, nil
}
