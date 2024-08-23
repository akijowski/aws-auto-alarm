package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type JSONCmd struct {
	inputs   []*cloudwatch.PutMetricAlarmInput
	isDelete bool
	writer   io.Writer
}

func (b *Builder) NewJSONCmd(wr io.Writer) *JSONCmd {
	return &JSONCmd{
		inputs:   b.inputs,
		isDelete: b.config.Delete,
		writer:   wr,
	}
}

func (j *JSONCmd) Execute(_ context.Context) error {
	// refactor duplicates
	// use json.Encoder
	if j.isDelete {
		b, err := json.Marshal(deleteInput(j.inputs))
		if err != nil {
			return fmt.Errorf("json error: %w", err)
		}

		_, err = j.writer.Write(b)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}

	} else {
		b, err := json.Marshal(j.inputs)
		if err != nil {
			return fmt.Errorf("json error: %w", err)
		}

		_, err = j.writer.Write(b)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}

	return nil
}
