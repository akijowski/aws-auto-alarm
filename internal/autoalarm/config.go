package autoalarm

import (
	"context"
	"errors"
	"fmt"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config is parsed data from flags, variables, or files.
type Config struct {
	Quiet        bool           `json:"quiet"`
	DryRun       bool           `json:"dryRun"`
	AlarmPrefix  string         `json:"alarmPrefix"`
	ARN          string         `json:"arn"`
	Delete       bool           `json:"delete"`
	OKActions    []string       `json:"okActions"`
	AlarmActions []string       `json:"alarmActions"`
	Overrides    map[string]any `json:"overrides"`
	ParsedARN    awsarn.ARN
}

func NewConfig(ctx context.Context) *Config {
	config, err := initConfiguration()
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Send()
	}

	if err := parseARN(config); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Send()
	}

	return config
}

func initConfiguration() (*Config, error) {
	pflag.StringP("file", "f", "", "read command options from a file")
	pflag.BoolP("quiet", "q", false, "set to only log errors")

	pflag.Parse()

	viper.SetEnvPrefix("AWS_AUTO_ALARM")
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()

	if viper.IsSet("file") {
		viper.SetConfigFile(viper.GetString("file"))
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := new(Config)

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return config, nil
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
