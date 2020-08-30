package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/exectree"
)

func Run(ctx context.Context, funcs ...interface{}) {
	tree, err := cmdtree.New(funcs)
	if err != nil {
		panic(err)
	}

	if err := exectree.Exec(ctx, tree, os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
