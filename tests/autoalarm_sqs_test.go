package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

type ExecutionHistory struct {
	ID                        int64                     `json:"id"`
	PreviousID                int64                     `json:"previous_id"`
	Timestamp                 time.Time                 `json:"timestamp"`
	Type                      string                    `json:"type"`
	ExecutionStartedDetails   ExecutionStartedDetails   `json:"execution_started_details,omitempty"`
	ExecutionSucceededDetails ExecutionSucceededDetails `json:"execution_succeeded_details,omitempty"`
	ExecutionFailedDetails    ExecutionFailedDetails    `json:"execution_failed_details,omitempty"`
}

type ExecutionStartedDetails struct {
	Input   string `json:"input"`
	RoleArn string `json:"role_arn"`
}

type ExecutionFailedDetails struct {
	Cause string `json:"cause"`
	Error string `json:"error"`
}

type ExecutionSucceededDetails struct {
	Output string `json:"output"`
}

func TestSQS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	f, err := os.ReadFile("../statefiles/autoalarm_sqs.asl.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("file contents: %s\n", f)

	uploadInput := &sfn.CreateStateMachineInput{
		Definition: aws.String(string(f)),
		Name:       aws.String("SQSTest"),
		RoleArn:    aws.String(IAMRoleDummy),
	}

	uploadOutput, err := sfnClient.CreateStateMachine(ctx, uploadInput)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupMachine(ctx, t, *uploadOutput.StateMachineArn)
	})

	machineARN := *uploadOutput.StateMachineArn

	t.Logf("machine ARN: %s\n", machineARN)

	runInput := &sfn.StartExecutionInput{
		StateMachineArn: uploadOutput.StateMachineArn,
		Input:           aws.String("{}"),
	}

	runOutput, err := sfnClient.StartExecution(ctx, runInput)
	if err != nil {
		t.Fatal(err)
	}

	resultOutput, err := pollForExecutionHistory(ctx, t, 5*time.Second, runOutput.ExecutionArn)
	if err != nil {
		t.Fatal(err)
	}

	out, err := json.MarshalIndent(mapExecutionHistory(resultOutput.Events), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\n", out)
}

// TODO: better abstraction of SFN Event types: Start, Execution, Result, etc
func mapExecutionHistory(events []types.HistoryEvent) []ExecutionHistory {
	var out []ExecutionHistory
	for _, event := range events {
		started := ExecutionStartedDetails{}
		succeeded := ExecutionSucceededDetails{}
		failed := ExecutionFailedDetails{}

		if event.ExecutionStartedEventDetails != nil {
			started = ExecutionStartedDetails{
				Input:   *event.ExecutionStartedEventDetails.Input,
				RoleArn: *event.ExecutionStartedEventDetails.RoleArn,
			}
		}

		if event.ExecutionSucceededEventDetails != nil {
			succeeded = ExecutionSucceededDetails{
				Output: *event.ExecutionSucceededEventDetails.Output,
			}
		}

		if event.ExecutionFailedEventDetails != nil {
			failed = ExecutionFailedDetails{
				Cause: *event.ExecutionFailedEventDetails.Cause,
				Error: *event.ExecutionFailedEventDetails.Error,
			}
		}

		out = append(out, ExecutionHistory{
			ID:                        event.Id,
			PreviousID:                event.PreviousEventId,
			Timestamp:                 *event.Timestamp,
			Type:                      string(event.Type),
			ExecutionStartedDetails:   started,
			ExecutionSucceededDetails: succeeded,
			ExecutionFailedDetails:    failed,
		})
	}

	return out
}
