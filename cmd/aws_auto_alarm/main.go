package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/awsclient"
	"github.com/akijowski/aws-auto-alarm/internal/cli"
)

func main() {
	ctx := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().WithContext(context.Background())

	pflag.StringP("file", "f", "", "read command options from a file")
	pflag.BoolP("quiet", "q", false, "set to only log errors")

	pflag.Parse()

	config := autoalarm.NewCLIConfig(ctx, pflag.CommandLine)
	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	ctx = zerolog.Ctx(ctx).With().Str("arn", config.ParsedARN.String()).Logger().WithContext(ctx)
	log := zerolog.Ctx(ctx)

	cw, err := awsclient.CloudWatch(ctx)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	c := cli.New(config, cw)
	if err := c.Run(ctx, os.Stdout); err != nil {
		log.Fatal().Err(err).Send()
	}
}
