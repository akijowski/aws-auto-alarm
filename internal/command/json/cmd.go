package json

import (
	"context"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog/log"
)

type CreateCmd struct {
	inputs []*cloudwatch.PutMetricAlarmInput
	wr     io.Writer
}

func NewCreateCmd(inputs []*cloudwatch.PutMetricAlarmInput, wr io.Writer) *CreateCmd {
	return &CreateCmd{
		inputs: inputs,
		wr:     wr,
	}
}

type DeleteCmd struct {
	input *cloudwatch.DeleteAlarmsInput
	wr    io.Writer
}

func NewDeleteCmd(input *cloudwatch.DeleteAlarmsInput, wr io.Writer) *DeleteCmd {
	return &DeleteCmd{
		input: input,
		wr:    wr,
	}
}

func (c *CreateCmd) Execute(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Debug().Msg("writing output as JSON")

	encoder := json.NewEncoder(c.wr)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c.inputs); err != nil {
		return err
	}

	return nil
}

func (d *DeleteCmd) Execute(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Debug().Msg("writing output as JSON")

	encoder := json.NewEncoder(d.wr)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(d.input); err != nil {
		return err
	}

	return nil
}
