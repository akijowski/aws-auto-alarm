package autoalarm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_loadViperConfig(t *testing.T) {
	//t.Parallel()

	defaultARNString := "arn:aws:cloudwatch:us-west-2:123456789012:alarm/my-alarm"
	defaultARN, _ := arn.Parse(defaultARNString)

	cases := map[string]struct {
		reader   func(t testing.TB) io.Reader
		fileType string
		cfg      func(t testing.TB) *Config
		wantCfg  *Config
		wantErr  bool
	}{
		"json file returns config": {
			reader: func(t testing.TB) io.Reader {
				json := `{"dryRun": true, "overrides": {"foo": "bar"}}`
				return strings.NewReader(json)
			},
			fileType: "json",
			cfg: func(t testing.TB) *Config {
				return &Config{
					ARN: defaultARNString,
				}
			},
			wantCfg: &Config{
				DryRun:    true,
				Overrides: map[string]any{"foo": "bar"},
				ARN:       defaultARNString,
				ParsedARN: defaultARN,
			},
		},
		"dotenv file returns config": {
			reader: func(t testing.TB) io.Reader {
				line1 := "QUIET=true\n"
				line2 := "OKACTIONS=arn:aws:sns:us-west-2:123456789012:my-topic,arn:aws:sns:us-east-1:123456789012:my-other-topic\n"
				line3 := "OVERRIDES={\"foo\":\"bar\",\"baz\":\"qux\"}\n"

				dotenv := fmt.Sprintf("%s%s%s", line1, line2, line3)
				return strings.NewReader(dotenv)
			},
			fileType: "dotenv",
			cfg: func(t testing.TB) *Config {
				return &Config{
					ARN: defaultARNString,
				}
			},
			wantCfg: &Config{
				Quiet: true,
				OKActions: []string{
					"arn:aws:sns:us-west-2:123456789012:my-topic",
					"arn:aws:sns:us-east-1:123456789012:my-other-topic",
				},
				Overrides: map[string]any{
					"foo": "bar",
					"baz": "qux",
				},
				ARN:       defaultARNString,
				ParsedARN: defaultARN,
			},
		},
		"config unmarshall failure returns error": {
			reader: func(t testing.TB) io.Reader {
				return bytes.NewReader([]byte(`{"dryRun": true}`))
			},
			fileType: "json",
			cfg: func(t testing.TB) *Config {
				return nil
			},
			wantErr: true,
		},
		"unsupported file type returns error": {
			reader: func(t testing.TB) io.Reader {
				return bytes.NewReader([]byte(`{"dryRun": true}`))
			},
			fileType: "unsupported",
			cfg: func(t testing.TB) *Config {
				return &Config{}
			},
			wantErr: true,
		},
		"invalid file returns error": {
			reader: func(t testing.TB) io.Reader {
				return bytes.NewReader([]byte("invalid"))
			},
			fileType: "json",
			cfg: func(t testing.TB) *Config {
				return &Config{}
			},
			wantErr: true,
		},
		"missing ARN returns error": {
			reader: func(t testing.TB) io.Reader {
				return bytes.NewReader([]byte(`{"dryRun": true}`))
			},
			fileType: "json",
			cfg: func(t testing.TB) *Config {
				return &Config{}
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		//tc := tc
		t.Run(name, func(t *testing.T) {
			//t.Parallel()

			viper.Reset()

			cfg := tc.cfg(t)
			err := loadViperConfig(tc.reader(t), tc.fileType, cfg)

			if tc.wantErr {
				assert.Error(t, err)
				t.Logf("actual error: %v", err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantCfg, cfg)
			}
		})
	}
}

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
