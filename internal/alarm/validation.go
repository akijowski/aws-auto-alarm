package alarm

import (
	"fmt"
	"strings"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
)

// IsValid will run the provided ARN through validation functions for each AWS service, returning the first error encountered.
func IsValid(arn awsarn.ARN) error {
	return validate(arn, eventBridgeValidator())
}

func validate(arn awsarn.ARN, validFuncs ...func(awsarn.ARN) error) error {
	for _, f := range validFuncs {
		if err := f(arn); err != nil {
			return err
		}
	}

	return nil
}

func eventBridgeValidator() func(awsarn.ARN) error {
	return func(a awsarn.ARN) error {
		// valid ARNs https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-manage-iam-access.html#eb-arn-format
		if a.Service != "events" {
			return nil
		}
		resourceType := strings.SplitN(a.Resource, "/", 2)[0]
		if resourceType != "rule" {
			return fmt.Errorf("eventbridge resource %s is not supported", resourceType)
		}

		return nil
	}
}
