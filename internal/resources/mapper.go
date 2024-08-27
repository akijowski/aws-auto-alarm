package resources

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

type resourceMapFn func(cfg *autoalarm.Config, m map[string]any)

type Mapper struct {
	cfg       *autoalarm.Config
	resources map[string]any
	fns       []resourceMapFn
}

func NewMapper(cfg *autoalarm.Config) *Mapper {
	resources := make(map[string]any)
	fns := []resourceMapFn{
		sqsResources,
	}

	return &Mapper{
		cfg:       cfg,
		resources: resources,
		fns:       fns,
	}
}

func (m *Mapper) Map(ctx context.Context) map[string]any {
	zerolog.Ctx(ctx).Debug().
		Int("functions_length", len(m.fns)).
		Msg("Mapping resources")
	for _, fn := range m.fns {
		fn(m.cfg, m.resources)
	}

	return m.resources
}
