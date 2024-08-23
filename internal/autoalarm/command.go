package autoalarm

import "context"

type Command interface {
	Execute(context.Context) error
}

/*
1. parse config
2. if dry run create external writer command
3. if live create cw writer command
4. command creation -> parse ARN, generate templates, generate cloudwatch alarm inputs
5. execute commands -> if delete use names
*/
