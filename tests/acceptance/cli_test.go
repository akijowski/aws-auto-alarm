package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/akijowski/aws-auto-alarm/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		service string
	}{
		{
			service: "sqs",
		},
		{
			service: "sqs_delete",
		},
		{
			service: "sqs_override_dlq",
		},
	}

	for _, tc := range cases {
		t.Run(tc.service, func(t *testing.T) {
			t.Parallel()

			tc := tc

			require := require.New(t)

			file, err := os.ReadFile(fmt.Sprintf("./fixtures/input/%s.json", tc.service))
			require.NoError(err)

			config := new(cli.Config)
			err = json.Unmarshal(file, &config)
			require.NoError(err)

			buf := new(bytes.Buffer)

			cli.Run(context.Background(), config, buf)

			wanted, err := os.ReadFile(fmt.Sprintf("./fixtures/output/%s.json", tc.service))
			require.NoError(err)

			assert.JSONEq(t, string(wanted), buf.String())
		})
	}

}
