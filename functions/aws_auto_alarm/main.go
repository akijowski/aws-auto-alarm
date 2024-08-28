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

	cw, err := awsclient.CloudWatch(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("Failed to create CloudWatch client")
	}
	handler := &task.AlarmHandler{
		MetricAPI: cw,
	}
	lambda.StartWithOptions(handler.Handle, lambda.WithContext(ctx))
}
