package main

import (
	"context"

	"github.com/akijowski/aws-auto-alarm/internal/task"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx := log.Logger.WithContext(context.Background())
	handler := &task.AlarmHandler{}
	lambda.StartWithOptions(handler.Handle, lambda.WithContext(ctx))
}
