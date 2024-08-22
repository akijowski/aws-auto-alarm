package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/rs/zerolog/log"
)

type CreateJSON struct {
	baseAlarm *cloudwatch.PutMetricAlarmInput
	writer    io.Writer
	tmpls     []*template.Template
}

func (c *CreateJSON) Execute(ctx context.Context) error {
	log.Ctx(ctx).Debug().Msg("executing command")

	buf := new(bytes.Buffer)

	for _, tmpl := range c.tmpls {
		b, err := templateAlarm(tmpl, nil, c.baseAlarm)
		if err != nil {
			return err
		}
		_, _ = buf.Write(b)
	}

	_, err := c.writer.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("unable to write data: %w", err)
	}

	return nil
}

func templateAlarm(t *template.Template, opt any, base *cloudwatch.PutMetricAlarmInput) ([]byte, error) {
	buf := new(bytes.Buffer)

	input := new(cloudwatch.PutMetricAlarmInput)
	copyAlarmBase(base, input)

	if err := t.Execute(buf, opt); err != nil {
		return nil, fmt.Errorf("unable to template alarm: %w", err)
	}

	if err := json.Unmarshal(buf.Bytes(), input); err != nil {
		return nil, fmt.Errorf("unable to parse json: %w", err)
	}

	buf.Reset()

	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to bytes: %w", err)
	}

	return bytes, nil
}

func copyAlarmBase(src, dest *cloudwatch.PutMetricAlarmInput) {
	dest.ActionsEnabled = src.ActionsEnabled
	dest.AlarmActions = src.AlarmActions
	dest.OKActions = src.OKActions
	dest.Tags = src.Tags
}
