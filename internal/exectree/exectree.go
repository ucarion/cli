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
	// The argparser module does most of the work of understanding what each arg
	// does to the tree.
	parser := argparser.New(tree)
	for _, arg := range args {
		if err := parser.ParseArg(arg); err != nil {
			return err
		}
	}

	// If the user passed a help flag or is invoking a command that isn't itself
	// executable, then show a help message.
	//
	// We do this check before the NoMoreArgs check because we want to let users
	// pass --help without necessarily making a correct invocation.
	if parser.ShowHelp || !parser.CommandTree.Func.IsValid() {
		_, err := HelpWriter.Write([]byte(cmdhelp.Help(parser.CommandTree, parser.Name)))
		return err
	}

	// Make sure that after all the args are passed, that we're in a valid
	// parsing state to leave things on.
	if err := parser.NoMoreArgs(); err != nil {
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
