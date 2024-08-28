package autoalarm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// overrideConfig allows for unmarshalling of a map[string]any
// https://github.com/spf13/viper/issues/523
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
	Version        int64             `json:"version"`
	Tags           map[string]string `json:"tags"`
}

func NewCLIConfig(ctx context.Context, pflags *pflag.FlagSet) *Config {
	log := zerolog.Ctx(ctx)

	if err := viper.BindPFlags(pflags); err != nil {
		log.Fatal().Err(err).Send()
	}

	config := new(Config)

	if viper.IsSet("file") {
		file, err := os.Open(viper.GetString("file"))
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		ext := filepath.Ext(file.Name())
		if err = loadViperConfig(file, ext[1:], config); err != nil {
			log.Fatal().Err(err).Send()
		}
	} else {
		log.Warn().Msg("configuration other than file for CLI is not well supported")
		if err := viper.Unmarshal(config); err != nil {
			log.Fatal().Err(err).Send()
		}
	}

	return config
}

func NewLambdaConfig(ctx context.Context, event *events.EventBridgeEvent) (*Config, error) {
	log := zerolog.Ctx(ctx)

	log.Debug().Msg("parsing config from event")
	config := new(Config)

	if err := loadEvent(ctx, event, config); err != nil {
		return nil, fmt.Errorf("unable to create lambda config: %w", err)
	}

	return config, nil
}

func loadEvent(ctx context.Context, event *events.EventBridgeEvent, config *Config) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("loading event")
	// TODO: instead of the event use the tag detail?
	detail := new(tagChangeDetail)
	if err := json.Unmarshal(event.Detail, detail); err != nil {
		return fmt.Errorf("unable to unmarshal detail: %w", err)
	}

	if isDelete(detail) {
		config.Delete = true
	}

	config.ARN = event.Resources[0]

	dotenv := mapToDotEnv(detail.Tags)

	if err := loadViperConfig(dotenv, "dotenv", config); err != nil {
		return fmt.Errorf("unable to load config from dotenv: %w", err)
	}

	return nil
}

func mapToDotEnv(env map[string]string) io.Reader {
	buf := new(bytes.Buffer)
	filteredMap := make(map[string]string)
	for k, v := range env {
		if nk, ok := strings.CutPrefix(k, "AWS_AUTO_ALARM_"); ok && k != "AWS_AUTO_ALARM_MANAGED" {
			filteredMap[nk] = v
		}
	}

	for k, v := range filteredMap {
		buf.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return buf
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

func isDelete(detail *tagChangeDetail) bool {
	del := false
	_, hasTag := detail.Tags["AWS_AUTO_ALARM_MANAGED"]
	if slices.Contains(detail.ChangedTagKeys, "AWS_AUTO_ALARM_MANAGED") && !hasTag {
		del = true
	}

	return del
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
