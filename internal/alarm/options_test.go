package alarm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithOKActions(t *testing.T) {
	t.Parallel()

	wanted := []string{"arn1", "arn2"}

	d := new(Data)

	err := WithOKActions(wanted...)(context.TODO(), d)

	assert := assert.New(t)

	assert.NoError(err)
	assert.EqualValues(wanted, d.OKActions)
}

func TestWithAlarmActions(t *testing.T) {
	t.Parallel()

	wanted := []string{"arn1", "arn2"}

	d := new(Data)

	err := WithAlarmActions(wanted...)(context.TODO(), d)

	assert := assert.New(t)

	assert.NoError(err)
	assert.EqualValues(wanted, d.AlarmActions)
}
