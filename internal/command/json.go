package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog"
)

type JSONCmd struct {
	inputs      []*cloudwatch.PutMetricAlarmInput
	isDelete    bool
	prettyPrint bool
	writer      io.Writer
}

func (b *Builder) NewJSONCmd(wr io.Writer) *JSONCmd {
	return &JSONCmd{
		inputs:      b.inputs,
		isDelete:    b.config.Delete,
		prettyPrint: b.config.PrettyPrint,
		writer:      wr,
	}
}

func (j *JSONCmd) Execute(ctx context.Context) error {
	zerolog.Ctx(ctx).Debug().Msg("writing output as JSON")
	if j.inputs == nil {
		return errors.New("no inputs provided")
	}
	if j.writer == nil {
		return errors.New("no writer provided")
	}

	encoder := json.NewEncoder(j.writer)
	if j.prettyPrint {
		encoder.SetIndent("", "  ")
	}
	var err error
	if j.isDelete {
		err = encoder.Encode(deleteInput(j.inputs))
	} else {
		err = encoder.Encode(j.inputs)
	}
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	return nil
}
