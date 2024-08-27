package resources

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

func tagResources(cfg *autoalarm.Config, m map[string]any) {
	tags := make([]types.Tag, 0)
	for k, v := range cfg.Tags {
		tag := types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		tags = append(tags, tag)
	}
	m["Tags"] = tags
}
