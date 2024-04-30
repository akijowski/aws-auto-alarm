package main

import (
	"context"
	"flag"
	"os"

	"github.com/akijowski/aws-auto-alarm/internal/alarm"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/rs/zerolog"
)

var (
	flagDryRun = flag.Bool("dry-run", false, "whether to update the cloudwatch alarms")
)

func main() {
	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	flag.Parse()

	ctx := context.Background()

	if len(os.Args) < 2 {
		log.Fatal().Msg("missing ARN as an argument")
	}
	arn := flag.Arg(0)
	if !awsarn.IsARN(arn) {
		log.Fatal().Str("arg", arn).Msg("ARN is not valid")
	}

	ctx = log.With().Str("arn", arn).Logger().WithContext(ctx)

	opts := []alarm.DataOptionFunc{
		alarm.AddServiceData,
		alarm.WithOKActions("arn:aws:sns:us-east-2:0123456789012:topic/foo"),
	}

	data, err := alarm.FromARN(ctx, arn, opts...)
	if err != nil {
		log.Fatal().Ctx(ctx).Err(err).Send()
	}

	if *flagDryRun {
		err = alarm.WithData(data)(os.Stdout)
	} else {
		err = alarm.WriteToCloudwatch(ctx, nil, alarm.WithData(data))
	}

	if err != nil {
		log.Fatal().Ctx(ctx).Err(err).Send()
	}
}
