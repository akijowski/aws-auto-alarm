package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/akijowski/aws-auto-alarm/internal/alarm"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
)

var (
	flagFile = flag.String("file", "", "read command options from a file")
)

type cmdConfig struct {
	Quiet        bool     `json:"quiet"`
	IsDryRun     bool     `json:"isDryRun"`
	AlarmPrefix  string   `json:"alarmPrefix"`
	ARN          string   `json:"arn"`
	IsDelete     bool     `json:"delete"`
	OKActions    []string `json:"okActions"`
	AlarmActions []string `json:"alarmActions"`
}

func main() {
	flag.Parse()

	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	config := new(cmdConfig)

	if *flagFile != "" {
		configFile, err := os.ReadFile(*flagFile)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to open file")
		}

		err = json.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to parse configuration")
		}
	}

	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	log.Debug().Interface("config", config).Msg("configuration created")

	arn, err := awsarn.Parse(config.ARN)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse ARN")
	}

	log.Debug().Interface("arn", arn).Msg("parsed ARN")

	if config.IsDryRun {
		write, err := alarm.NewWriter(arn, config.IsDelete, with(config))
		if err != nil {
			log.Fatal().Err(err).Msg("unable to create writer")
		}

		err = write(os.Stdout)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	} else {
		err = alarm.UpdateCloudwatch(context.TODO(), nil, arn, config.IsDelete, with(config))
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}
}

func with(config *cmdConfig) func(o *alarm.Options) {
	return func(o *alarm.Options) {
		o.AlarmActions = config.AlarmActions
		o.AlarmPrefix = config.AlarmPrefix
	}
}
