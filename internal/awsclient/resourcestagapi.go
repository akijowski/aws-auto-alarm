package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
)

func ResourcesTagAPI(ctx context.Context) (*resourcegroupstaggingapi.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return resourcegroupstaggingapi.NewFromConfig(cfg), nil
}
