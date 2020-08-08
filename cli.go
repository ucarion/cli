package cli

import (
	"context"
	"os"

	"github.com/ucarion/cli/internal/cmd"
)

func Run(ctx context.Context, cmds ...interface{}) error {
	c, err := cmd.FromFunc(cmds[0])
	if err != nil {
		panic(err)
	}

	return c.Exec(ctx, os.Args[1:])
}
