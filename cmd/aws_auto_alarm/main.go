package main

import (
	"context"
	"os"

	"github.com/akijowski/aws-auto-alarm/internal/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	ctx := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().WithContext(context.Background())

	config := initConfig(ctx)

	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	log.Ctx(ctx).Debug().Interface("config", config).Msg("configuration created")

	cli.Run(ctx, config, os.Stdout)
}

func initConfig(ctx context.Context) *cli.Config {
	pflag.StringP("file", "f", "", "read command options from a file")
	pflag.BoolP("quiet", "q", false, "set to only log errors")

	pflag.Parse()

	viper.SetEnvPrefix("AWS_AUTO_ALARM")
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()

	if viper.IsSet("file") {
		viper.SetConfigFile(viper.GetString("file"))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Send()
	}

	config := new(cli.Config)

	if err := viper.Unmarshal(&config); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Send()
	}

	return config
}
