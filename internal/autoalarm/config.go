package autoalarm

import (
	"context"
	"errors"
	"fmt"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config is parsed data from flags, variables, or files.
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

func NewCLIConfig(ctx context.Context, pflags *pflag.FlagSet) *Config {
	log := zerolog.Ctx(ctx)
	setEnv()

	if err := viper.BindPFlags(pflags); err != nil {
		log.Fatal().Err(err).Send()
	}

	if viper.IsSet("file") {
		viper.SetConfigFile(viper.GetString("file"))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("unable to read config from file")
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read config")
	}

	return config
}

func setEnv() {
	viper.SetEnvPrefix("AWS_AUTO_ALARM")
	viper.AutomaticEnv()
}

func loadConfig() (*Config, error) {
	config := new(Config)

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := parseARN(config); err != nil {
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
