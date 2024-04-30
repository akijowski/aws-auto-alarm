package alarm

import "context"

func WithOKActions(arns ...string) DataOptionFunc {
	return func(_ context.Context, d *Data) error {
		d.OKActions = arns

		return nil
	}
}

func WithAlarmActions(arns ...string) DataOptionFunc {
	return func(_ context.Context, d *Data) error {
		d.AlarmActions = arns

		return nil
	}
}
