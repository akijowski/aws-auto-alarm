package alarm

// Options are configuration options applied to all generated alarms
type Options struct {
	AlarmPrefix  string
	OKActions    []string
	AlarmActions []string
}

func newOptions(opts ...func(o *Options)) *Options {
	o := new(Options)

	for _, f := range opts {
		f(o)
	}

	return o
}
