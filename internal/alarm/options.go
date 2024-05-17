package alarm

import "github.com/aws/aws-sdk-go-v2/aws/arn"

// Options are configuration options applied to all generated alarms.
type Options struct {
	// ARN is the ARN of the resource that will receive alarms
	ARN arn.ARN
	// AlarmPrefix is an optional prefix to an Alarm name
	AlarmPrefix string
	// OKActions is a slice of ARNs to trigger when an Alarm reaches an OK state
	OKActions []string
	// AlarmActions is a slice of ARNs to trigger when an Alarm reaches an ALARM state
	AlarmActions []string
	// Overrides is a map of value overrides that can be added to a created alarm.
	// Each service can specify how to use this map to generate the final template data.
	Overrides map[string]any
}

func newOptions(opts ...func(o *Options)) *Options {
	o := new(Options)

	for _, f := range opts {
		f(o)
	}

	return o
}
