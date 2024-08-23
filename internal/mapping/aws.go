package mapping

import (
	"context"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type resourceMapFn func(cfg *autoalarm.Config, m map[string]any)

type ResourceMapper struct {
	cfg       *autoalarm.Config
	resources map[string]any
	fns       []resourceMapFn
}

func NewResources(cfg *autoalarm.Config) *ResourceMapper {
	resources := make(map[string]any)
	fns := []resourceMapFn{
		sqsResources,
	}

	return &ResourceMapper{
		cfg:       cfg,
		resources: resources,
		fns:       fns,
	}
}

func (m *ResourceMapper) Map(ctx context.Context) map[string]any {
	for _, fn := range m.fns {
		fn(m.cfg, m.resources)
	}

	return m.resources
}
