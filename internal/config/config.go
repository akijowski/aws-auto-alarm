package config

import (
	"errors"
	"fmt"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
)

// Config is parsed data from flags, variables, or files. Well, just a file in this case.
type Config struct {
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

func ParseARN(cfg *Config) error {
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
