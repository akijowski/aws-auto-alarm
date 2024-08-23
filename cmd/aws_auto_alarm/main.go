package main

import (
	"context"
	"os"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/awsclient"
	"github.com/akijowski/aws-auto-alarm/internal/cli"
	"github.com/akijowski/aws-auto-alarm/internal/command"
	"github.com/rs/zerolog"
)

func main() {
	ctx := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().WithContext(context.Background())

	config := autoalarm.NewConfig(ctx)

	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	builder, err := command.DefaultBuilder(ctx, config)
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Send()
	}

	var cmd autoalarm.Command
	if config.DryRun {
		cmd = builder.NewJSONCmd(os.Stdout)
	} else {
		cw, err := awsclient.CloudWatch(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Fatal().Err(err).Send()
		}
		cmd = builder.NewCWCmd(cw)
	}

	cli.Run(ctx, cmd)
}
