// Package cli is a simple encapsulation of the cli program.
// A Run function is available to use with an io.Writer.
// I could have used a framework such as spf13/cobra, but that currently feels like overkill.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/akijowski/aws-auto-alarm/internal/command"
	"github.com/akijowski/aws-auto-alarm/internal/config"
	"github.com/akijowski/aws-auto-alarm/internal/template"
)

type DeleteCmdRegistry interface {
	DeleteCommand(ctx context.Context, cmdType string, finder command.AlarmNameFinder) (autoalarm.Command, error)
}

type CreateCmdRegistry interface {
	CreateCommand(ctx context.Context, cmdType string, loader command.AlarmLoader) (autoalarm.Command, error)
}

type CmdRegistry interface {
	CreateCmdRegistry
	DeleteCmdRegistry
}

type CLI struct {
	cfg  *config.Config
	cmds CmdRegistry
}

func New(cfg *config.Config, api autoalarm.MetricAlarmAPI, wr io.Writer) *CLI {
	return &CLI{
		cfg:  cfg,
		cmds: command.DefaultRegistry(api, wr),
	}
}

func (c *CLI) Run(ctx context.Context) error {
	log.Ctx(ctx).
		Info().
		Interface("config", c.cfg).
		Msg("running cli")
	cmdType := "cloudwatch"
	if c.cfg.DryRun {
		cmdType = "json"
	}

	cmd, err := c.cmds.CreateCommand(ctx, cmdType, template.NewFileLoader(ctx, c.cfg))
	if c.cfg.Delete {
		cmd, err = c.cmds.DeleteCommand(ctx, cmdType, template.NewFileFinder(ctx, c.cfg))
	}
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	return cmd.Execute(ctx)
}
