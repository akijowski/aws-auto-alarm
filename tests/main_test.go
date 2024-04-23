package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

var (
	sfnClient *sfn.Client
)

const (
	IAMRoleDummy = "arn:aws:iam::012345678901:role/DummyRole"
)

func TestMain(m *testing.M) {
	configOpts := []func(*config.LoadOptions) error{
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("FOO", "BAR", "SESS")),
	}
	config, err := config.LoadDefaultConfig(context.Background(), configOpts...)
	if err != nil {
		fmt.Printf("Error with config: %v", err)
		panic(err)
	}

	client := sfn.NewFromConfig(config, func(o *sfn.Options) {
		// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/#v2-endpointresolverv2--baseendpoint
		o.BaseEndpoint = aws.String("http://localhost:8083")
	})
	sfnClient = client

	m.Run()
}

func cleanupMachine(ctx context.Context, tb testing.TB, arn string) {
	_, err := sfnClient.DeleteStateMachine(ctx, &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(arn),
	})
	if err != nil {
		tb.Error(err)
	}
	tb.Logf("removed state machine: %s\n", arn)
}

func pollForExecutionHistory(ctx context.Context, tb testing.TB, timeout time.Duration, arn *string) (*sfn.GetExecutionHistoryOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout polling for execution history: %w", ctx.Err())
	default:
		out, err := sfnClient.GetExecutionHistory(ctx, &sfn.GetExecutionHistoryInput{ExecutionArn: arn})
		if err != nil {
			var dne *types.ExecutionDoesNotExist
			if errors.As(err, &dne) {
				tb.Logf("history does not exist, retrying.  ARN: %s\n", *arn)
			} else {
				return nil, err
			}
		}
		return out, err
	}
}
