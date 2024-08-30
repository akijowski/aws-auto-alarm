package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
)

// ResourcesTagAPI returns a client that represents the AWS Resource Groups and Tagging API.
// This API is useful for querying an AWS account for resources based on tag information.
// See the docs for a list of supported AWS services:
// https://docs.aws.amazon.com/resourcegroupstagging/latest/APIReference/supported-services.html
func ResourcesTagAPI(ctx context.Context) (*resourcegroupstaggingapi.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return resourcegroupstaggingapi.NewFromConfig(cfg), nil
}
