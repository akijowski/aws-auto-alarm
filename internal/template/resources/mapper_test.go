package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
)

func TestNewMapper(t *testing.T) {
	t.Parallel()

	cfg := &autoalarm.Config{}

	mapper := NewMapper(cfg)

	assert := assert.New(t)

	assert.Equal(cfg, mapper.cfg)
	assert.NotNil(mapper.resources)
	assert.GreaterOrEqual(len(mapper.fns), 1)
}

func TestMapper_Map(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	t.Run("calls map functions", func(t *testing.T) {
		t.Parallel()

		fnCalled := 0
		testFn := func(cfg *autoalarm.Config, m map[string]any) {
			fnCalled++
			m["test"] = "test"
		}
		mapper := &Mapper{
			cfg:       &autoalarm.Config{},
			resources: map[string]any{},
			fns: []resourceMapFn{
				testFn,
			},
		}
		_ = mapper.Map(context.TODO())

		assert.Equal(1, fnCalled)
	})

	t.Run("returns mapped resources", func(t *testing.T) {
		t.Parallel()

		testFn := func(cfg *autoalarm.Config, m map[string]any) {
			m["test"] = "test"
		}
		mapper := &Mapper{
			cfg:       &autoalarm.Config{},
			resources: map[string]any{},
			fns: []resourceMapFn{
				testFn,
			},
		}
		resources := mapper.Map(context.TODO())

		assert.Equal("test", resources["test"])
	})
}
