package command

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ErrorWriter struct {
	buf *bytes.Buffer
	err error
}

func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	w.buf.Write(p)
	return 0, w.err
}

func (w *ErrorWriter) Read(p []byte) (n int, err error) {
	return w.buf.Read(p)
}

func TestJSONCmd_Execute(t *testing.T) {
	t.Parallel()

	testBytes, err := json.Marshal(defaultTestInput(t))
	require.NoError(t, err)

	testDeleteBytes, err := json.Marshal(defaultTestDeleteInput(t))
	require.NoError(t, err)

	cases := map[string]struct {
		inputs    func(testing.TB) []*cloudwatch.PutMetricAlarmInput
		writer    func(testing.TB) io.ReadWriter
		isDelete  bool
		wantError bool
		wantBytes []byte
	}{
		"delete write errors are returned": {
			inputs: defaultTestInput,
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return &ErrorWriter{buf: new(bytes.Buffer), err: io.ErrShortWrite}
			},
			isDelete:  true,
			wantError: true,
		},
		"delete is returned as json": {
			inputs: defaultTestInput,
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return new(bytes.Buffer)
			},
			isDelete:  true,
			wantBytes: testDeleteBytes,
		},
		"nil writer returns an error": {
			inputs: defaultTestInput,
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return nil
			},
			wantError: true,
		},
		"nil inputs returns an error": {
			inputs: func(t testing.TB) []*cloudwatch.PutMetricAlarmInput {
				t.Helper()
				return nil
			},
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return new(bytes.Buffer)
			},
			wantError: true,
		},
		"write errors are returned": {
			inputs: defaultTestInput,
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return &ErrorWriter{buf: new(bytes.Buffer), err: io.ErrShortWrite}
			},
			wantError: true,
		},
		"inputs are returned as json": {
			inputs: defaultTestInput,
			writer: func(t testing.TB) io.ReadWriter {
				t.Helper()

				return new(bytes.Buffer)
			},
			wantBytes: testBytes,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buf := tc.writer(t)

			j := &JSONCmd{
				inputs:   tc.inputs(t),
				isDelete: tc.isDelete,
				writer:   buf,
			}

			err := j.Execute(context.TODO())
			assert.Equal(t, tc.wantError, err != nil)

			if !tc.wantError {
				b, err := io.ReadAll(buf)
				require.NoError(t, err)

				assert.JSONEq(t, string(tc.wantBytes), string(b))
			}
		})
	}
}

func defaultTestInput(t testing.TB) []*cloudwatch.PutMetricAlarmInput {
	t.Helper()
	return []*cloudwatch.PutMetricAlarmInput{
		{
			AlarmName: aws.String("test-alarm"),
		},
	}
}

func defaultTestDeleteInput(t testing.TB) *cloudwatch.DeleteAlarmsInput {
	t.Helper()
	names := make([]string, 0)
	for _, input := range defaultTestInput(t) {
		names = append(names, aws.ToString(input.AlarmName))
	}

	return &cloudwatch.DeleteAlarmsInput{
		AlarmNames: names,
	}
}
