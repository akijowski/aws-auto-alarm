// Package cli is a simple encapsulation of the cli program.
// A Run function is available to use with an autoalarm.Command.
// I could have used a framework such as spf13/cobra, but that currently feels like overkill.
package cli

import (
	"context"

	"github.com/akijowski/aws-auto-alarm/internal/autoalarm"
	"github.com/rs/zerolog"
)

// Run(ctx, <pkg>.Config, wr io.Writer) error
// newCommandFromConfig
// err = <pkg2>.execute(cmd)

// for 1 ARN
// to update:
// parse ARN
// parse templates for ARN.Service
// generate template input for Config
// generate cw api inputs from templates
// apply inputs to cw api
//
// to delete:
// parse ARN
// parse templates for ARN.Service
// generate template input for Config
// generate cw api inputs from templates
// reduce to alarm names
// apply names to cw api

// Run executes the provided autoalarm.Command. Any errors are set as Fatal and will exit 1.
func Run(ctx context.Context, cmd autoalarm.Command) {
	log := zerolog.Ctx(ctx)
	if err := cmd.Execute(ctx); err != nil {
		log.Fatal().Err(err).Send()
	}
}
