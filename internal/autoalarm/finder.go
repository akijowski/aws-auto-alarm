package autoalarm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

// sample ARNs: arn:aws:cloudwatch:us-east-2:123456789012:alarm:this alarm-name has spaces and / other < characters > to handle
// arn:aws:cloudwatch:us-east-2:123456789012:alarm:this-alarm-is-all-kebobs-and-numbers-like-123

type NameFinder struct {
	api GetResourcesAPI
	arn arn.ARN
}

func NewNameFinder(api GetResourcesAPI, arn arn.ARN) *NameFinder {
	return &NameFinder{
		api: api,
		arn: arn,
	}
}

func (f *NameFinder) Find(ctx context.Context) ([]string, error) {
	input := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    aws.String("AWS_AUTO_ALARM_MANAGED"),
				Values: []string{"true"},
			},
			{
				Key:    aws.String("AWS_AUTO_ALARM_SOURCE_ARN"),
				Values: []string{f.arn.String()},
			},
		},
	}

	output, err := f.api.GetResources(ctx, input)
	if err != nil {
		return nil, err
	}

	alarmNames := make([]string, 0)
	for _, mapping := range output.ResourceTagMappingList {
		alarmARN, err := arn.Parse(aws.ToString(mapping.ResourceARN))
		if err != nil {
			return nil, err
		}
		alarmNames = append(alarmNames, alarmARN.Resource)
	}

	return alarmNames, nil
}
