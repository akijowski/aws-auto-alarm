package autoalarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// overrideConfig allows for unmarshalling of a map[string]any
// https://github.com/spf13/viper/issues/523
// After replacing viper this may be removed
type overrideConfig map[string]any

func (c *overrideConfig) UnmarshalText(text []byte) error {
	var cfg map[string]any
	if err := json.Unmarshal(text, &cfg); err != nil {
		return err
	}
	*c = cfg
	return nil
}

func (c *overrideConfig) UnmarshalJSON(data []byte) error {
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	*c = cfg
	return nil
}

type tagConfig map[string]string

func (c *tagConfig) UnmarshalText(text []byte) error {
	var cfg map[string]string
	if err := json.Unmarshal(text, &cfg); err != nil {
		return err
	}
	*c = cfg
	return nil
}

func (c *tagConfig) UnmarshalJSON(data []byte) error {
	var cfg map[string]string
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	*c = cfg
	return nil
}

// Config is parsed data from flags, variables, or files.
type Config struct {
	Quiet        bool           `json:"quiet"`
	DryRun       bool           `json:"dryRun"`
	PrettyPrint  bool           `json:"prettyPrint"`
	AlarmPrefix  string         `json:"alarmPrefix"`
	ARN          string         `json:"arn"`
	Delete       bool           `json:"delete"`
	OKActions    []string       `json:"okActions"`
	AlarmActions []string       `json:"alarmActions"`
	Overrides    overrideConfig `json:"overrides"`
	Tags         tagConfig      `json:"tags"`
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

	if err := viper.BindPFlags(pflags); err != nil {
		logger.Fatal().Err(err).Send()
	}

	config := new(Config)

	if viper.IsSet("file") {
		file, err := os.Open(viper.GetString("file"))
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
		ext := filepath.Ext(file.Name())
		if err = loadViperConfig(file, ext[1:], config); err != nil {
			logger.Fatal().Err(err).Send()
		}
	} else {
		logger.Warn().Msg("configuration other than file for CLI is not well supported")
		if err := viper.Unmarshal(config); err != nil {
			logger.Fatal().Err(err).Send()
		}
	}

	return config
}

func loadViperConfig(r io.Reader, fileType string, cfg *Config) error {
	if !slices.Contains(viper.SupportedExts, fileType) {
		return fmt.Errorf("unsupported file type: %s", fileType)
	}
	viper.SetConfigType(fileType)
	if err := viper.MergeConfig(r); err != nil {
		return fmt.Errorf("unable to read config to viper: %w", err)
	}

	decode := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	)
	if err := viper.Unmarshal(cfg, decode); err != nil {
		return fmt.Errorf("unable to unmarshal config: %w", err)
	}

	if err := parseARN(cfg); err != nil {
		return fmt.Errorf("unable to parse ARN from config: %w", err)
	}

	return nil
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
