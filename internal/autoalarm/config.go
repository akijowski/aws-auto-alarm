package autoalarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

// Config is parsed data from flags, variables, or files. Well, just a file in this case.
type Config struct {
	Quiet        bool              `json:"quiet"`
	DryRun       bool              `json:"dryRun"`
	PrettyPrint  bool              `json:"prettyPrint"`
	AlarmPrefix  string            `json:"alarmPrefix"`
	ARN          string            `json:"arn"`
	Delete       bool              `json:"delete"`
	OKActions    []string          `json:"okActions"`
	AlarmActions []string          `json:"alarmActions"`
	Overrides    map[string]any    `json:"overrides"`
	Tags         map[string]string `json:"tags"`
	ParsedARN    awsarn.ARN
}

type tagChangeDetail struct {
	ChangedTagKeys []string          `json:"changed-tag-keys"`
	Service        string            `json:"service"`
	ResourceType   string            `json:"resource-type"`
	Version        float64           `json:"version"`
	Tags           map[string]string `json:"tags"`
}

func NewCLIConfig(ctx context.Context, pflags *pflag.FlagSet) *Config {
	logger := log.Ctx(ctx)

	config := new(Config)

	filePath, err := pflags.GetString("file")
	if err != nil {
		logger.Fatal().Err(fmt.Errorf("the flag file was not set: %w", err)).Send()
	}

	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	b, err := io.ReadAll(file)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	if err = json.Unmarshal(b, &config); err != nil {
		logger.Fatal().Err(err).Send()
	}

	if err = parseARN(config); err != nil {
		logger.Fatal().Err(err).Send()
	}

	return config
}

func parseARN(cfg *Config) error {
	if cfg.ARN == "" {
		return errors.New("ARN is required")
	}

	arn, err := awsarn.Parse(cfg.ARN)
	if err != nil {
		return fmt.Errorf("unable to parse ARN from config: %w", err)
	}

	cfg.ParsedARN = arn

	return nil
}
