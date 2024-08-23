package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

func CloudWatch(ctx context.Context) (*cloudwatch.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return cloudwatch.NewFromConfig(cfg), nil
}
