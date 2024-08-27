package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/task"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx := zerolog.New(os.Stdout).With().Timestamp().Logger().WithContext(context.Background())
	handler := &task.AlarmHandler{}
	lambda.StartWithOptions(handler.Handle, lambda.WithContext(ctx))
}
