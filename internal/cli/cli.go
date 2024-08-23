// Package cli is a simple encapsulation of the cli program.
// A Run function is available to use with an io.Writer.
// I could have used a framework such as spf13/cobra, but that currently feels like overkill.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/command"
)

type CLI struct {
	cfg *autoalarm.Config
	api command.MetricAlarmAPI
}

func New(cfg *autoalarm.Config, api command.MetricAlarmAPI) *CLI {
	return &CLI{
		cfg: cfg,
		api: api,
	}
}

func (c *CLI) Run(ctx context.Context, wr io.Writer) error {
	builder, err := command.DefaultBuilder(ctx, c.cfg)
	if err != nil {
		return fmt.Errorf("unable to build command: %w", err)
	}

	var cmd autoalarm.Command
	if c.cfg.DryRun {
		cmd = builder.NewJSONCmd(wr)
	} else {
		cmd = builder.NewCWCmd(c.api)
	}

	return cmd.Execute(ctx)
}
