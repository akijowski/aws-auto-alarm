package main

import (
	"context"
	"os"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/awsclient"
	"github.com/akijowski/aws-auto-alarm/internal/cli"
	"github.com/rs/zerolog"
)

func main() {
	ctx := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().WithContext(context.Background())

	config := autoalarm.NewConfig(ctx)
	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	log := zerolog.Ctx(ctx)

	cw, err := awsclient.CloudWatch(ctx)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	cli := cli.New(config, cw)
	if err := cli.Run(ctx, os.Stdout); err != nil {
		log.Fatal().Err(err).Send()
	}
}
