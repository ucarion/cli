package exectree

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/argparser"
	"github.com/ucarion/cli/internal/cmdhelp"
	"github.com/ucarion/cli/internal/cmdtree"
)

var HelpWriter io.Writer = os.Stdout

func Exec(ctx context.Context, tree cmdtree.CommandTree, args []string) error {
	parser := argparser.New(tree)
	for _, arg := range args {
		if err := parser.ParseArg(arg); err != nil {
			return err
		}
	}

	if err := parser.NoMoreArgs(); err != nil {
		return err
	}

	if parser.ShowHelp {
		_, err := HelpWriter.Write([]byte(cmdhelp.Help(parser.CommandTree, parser.Name)))
		return err
	}

	out := parser.CommandTree.Func.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		parser.Config,
	})

	err := out[0].Interface()
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", strings.Join(parser.Name, " "), err.(error))
}
