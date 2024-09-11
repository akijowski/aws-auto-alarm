package resources

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

type resourceMapFn func(cfg *config.Config, m map[string]any)

// Mapper contains functions to generate the map for alarmData.Resources.
type Mapper struct {
	cfg       *config.Config
	resources map[string]any
	fns       []resourceMapFn
}

func NewMapper(cfg *config.Config) *Mapper {
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
	log.Ctx(ctx).Debug().
		Int("functions_length", len(m.fns)).
		Bool("has_overrides", len(m.cfg.Overrides) > 0).
		Interface("overrides", m.cfg.Overrides).
		Msg("Mapping resources")
	for _, fn := range m.fns {
		fn(m.cfg, m.resources)
	}

	return m.resources
}
