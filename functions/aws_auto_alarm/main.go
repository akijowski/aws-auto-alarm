package main

import (
	"github.com/akijowski/aws-auto-alarm/internal/task"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	handler := &task.AlarmHandler{}
	lambda.Start(handler.Handle)
}
