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
	api autoalarm.MetricAlarmAPI
}

func New(cfg *autoalarm.Config, api autoalarm.MetricAlarmAPI) *CLI {
	return &CLI{
		cfg: cfg,
		api: api,
	}
}

func (c *CLI) Run(ctx context.Context, wr io.Writer) error {
	cmdFactory := command.DefaultFactory(ctx, c.cfg)
	var cmd autoalarm.Command
	var err error
	if c.cfg.DryRun {
		cmd, err = cmdFactory.WithWriter(wr)(ctx, c.cfg.Delete)
	} else {
		cmd, err = cmdFactory.WithMetricAPI(c.api)(ctx, c.cfg.Delete)
	}
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	return cmd.Execute(ctx)
}
