// Package cli is a simple encapsulation of the cli program.
// A Run function is available to use with a configuration and an interface for writing output.
// I could have used a framework such as spf13/cobra, but that currently feels like overkill.
package cli

import (
	"context"
	"io"

	"github.com/akijowski/aws-auto-alarm/internal/alarm"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
)

// Config is parsed CLI configuration data from flags, variables, or files.
type Config struct {
	Quiet        bool           `json:"quiet"`
	DryRun       bool           `json:"dryRun"`
	AlarmPrefix  string         `json:"alarmPrefix"`
	ARN          string         `json:"arn"`
	Delete       bool           `json:"delete"`
	OKActions    []string       `json:"okActions"`
	AlarmActions []string       `json:"alarmActions"`
	Overrides    map[string]any `json:"overrides"`
	parsedARN    awsarn.ARN
}

// Run executes the CLI using the provided Config.
// If Config.DryRun is true, program output will be sent to the provided io.Writer.
// Otherwise, an AWS client will be created.
// Setting Config.Delete will change the output to represent the data required to delete CloudWatch Alarms.
func Run(ctx context.Context, cfg *Config, wr io.Writer) {
	log := zerolog.Ctx(ctx)

	arn, err := awsarn.Parse(cfg.ARN)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse ARN")
	}

	cfg.parsedARN = arn

	log.Debug().Interface("arn", arn).Msg("parsed ARN")

	if err := alarm.IsValid(arn); err != nil {
		log.Fatal().Err(err).Msg("ARN is not supported")
	}

	if cfg.DryRun {
		writeTo, err := alarm.NewWriter(cfg.Delete, with(cfg))
		if err != nil {
			log.Fatal().Err(err).Msg("unable to create writer")
		}

		err = writeTo(wr)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	} else {
		err = alarm.UpdateCloudwatch(ctx, nil, cfg.Delete, with(cfg))
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}
}

func with(config *Config) func(o *alarm.Options) {
	return func(o *alarm.Options) {
		o.ARN = config.parsedARN
		o.AlarmActions = config.AlarmActions
		o.AlarmPrefix = config.AlarmPrefix
		o.Overrides = config.Overrides
	}
}
