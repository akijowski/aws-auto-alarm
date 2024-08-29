package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/awsclient"
	"github.com/akijowski/aws-auto-alarm/internal/task"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx := zerolog.New(os.Stdout).With().Timestamp().Logger().WithContext(context.Background())

	logLevel := os.Getenv("AWS_AUTO_ALARM_LOG_LEVEL")
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	cw, err := awsclient.CloudWatch(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("Failed to create CloudWatch client")
	}
	tag, err := awsclient.ResourcesTagAPI(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("Failed to create Resources Tag client")
	}
	handler := &task.AlarmHandler{
		MetricAPI:   cw,
		ResourceAPI: tag,
	}
	lambda.StartWithOptions(handler.Handle, lambda.WithContext(ctx))
}
