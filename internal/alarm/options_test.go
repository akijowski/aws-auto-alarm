package alarm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	expected := &Options{
		AlarmPrefix:  t.Name(),
		AlarmActions: []string{"test"},
	}

	actual := newOptions(func(o *Options) {
		o.AlarmPrefix = expected.AlarmPrefix
		o.AlarmActions = expected.AlarmActions
		o.OKActions = expected.OKActions
	})

	assert.Equal(t, expected, actual)
}
