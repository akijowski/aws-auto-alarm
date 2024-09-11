package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"

	"github.com/akijowski/aws-auto-alarm/internal/config"
)

func NewConfig(ctx context.Context, pflags *pflag.FlagSet) *config.Config {
	logger := log.Ctx(ctx)

	cfg := new(config.Config)

	filePath, err := pflags.GetString("file")
	if err != nil {
		logger.Fatal().Err(fmt.Errorf("the flag file was not set: %w", err)).Send()
	}

	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	b, err := io.ReadAll(file)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		logger.Fatal().Err(err).Send()
	}

	if err = config.ParseARN(cfg); err != nil {
		logger.Fatal().Err(err).Send()
	}

	return cfg
}
