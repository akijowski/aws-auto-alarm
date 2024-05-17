package alarm

import (
	"encoding/json"
	"fmt"
	"io"
)

type writeFunc func(wr io.Writer) error

func NewWriter(delete bool, opts ...func(o *Options)) (writeFunc, error) {
	opt := newOptions(opts...)

	alarms, err := generateAlarms(opt, embededTemplatesForARN(opt.ARN))
	if err != nil {
		return nil, err
	}

	if delete {
		return genericWriteFunc(toAlarmDeletionInput(alarms)), nil
	}

	return genericWriteFunc(alarms), nil
}

func genericWriteFunc(inputs any) writeFunc {
	return func(wr io.Writer) error {
		b, err := json.Marshal(inputs)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}

		_, err = wr.Write(b)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		return nil
	}
}
