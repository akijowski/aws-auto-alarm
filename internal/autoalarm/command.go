package autoalarm

import "context"

// Command is a generic interface that can be used by cli or lambda environments.
type Command interface {
	Execute(context.Context) error
}
