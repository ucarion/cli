package exectree

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/cmdtree"
)

func Exec(ctx context.Context, tree cmdtree.CommandTree, args []string) error {
	return exec(ctx, reflect.New(tree.Config).Elem(), tree, args)
}

func exec(ctx context.Context, config reflect.Value, tree cmdtree.CommandTree, args []string) error {
	// Process args until there are no more args.
	for len(args) > 0 {
		var arg string // the arg we're processing next
		arg, args = args[0], args[1:]

		// What kind of argument are we dealing with?
		switch {
		case strings.HasPrefix(arg, "-"):
			// It's one or more short flags in a bundle. A "bundle" is a set of
			// flags like "-abc", which is an alias for "-a -b -c", assuming
			// "-a" and "-b" are boolean flags that don't take a value.
			//
			// Let's now consume the chars in arg one-by-one, to handle each
			// bundled flag.
			chars := arg[1:] // strip out the leading "-"
			for len(chars) > 0 {
				var char string
				char, chars = string(chars[0]), chars[1:]

				// Try to find the corresponding short flag in the config.
				flag, _ := getShortFlag(tree, char) // TODO handle not there

				// Does the flag take a value?
				if flagTakesValue(tree.Config, flag) {
					fmt.Println("chars", flag, chars)
					// The flag does take a value. It may take on one of two
					// forms, which must be handled separately.
					if len(chars) > 0 {
						// The flag is in the "stuck" form, e.g. "-ojson". The flag's
						// value is the rest of the arg following the name of the flag.
						setConfigField(config, flag.Field, chars)
						chars = "" // reset chars so we stop looking for more flags
					} else {
						// The flag is in the "separate" form, e.g. "-o json". The next
						// arg is the flag's value.
						arg, args = args[0], args[1:] // TODO no next arg
						setConfigField(config, flag.Field, arg)
					}
				} else {
					// The flag does not take a value. We just assign the
					// appropriate field to "true".
					setConfigField(config, flag.Field, "")
				}
			}
		}
	}

	out := tree.Func.Call([]reflect.Value{reflect.ValueOf(ctx), config})

	err := out[0].Interface()
	if err == nil {
		return nil
	}

	return err.(error)
}

func getShortFlag(tree cmdtree.CommandTree, s string) (cmdtree.Flag, bool) {
	for _, f := range tree.Flags {
		for _, name := range f.ShortNames {
			if name == s {
				return f, true
			}
		}
	}

	return cmdtree.Flag{}, false
}

func setConfigField(config reflect.Value, index []int, val string) {
	// TODO other field types
	switch v := config.FieldByIndex(index).Addr().Interface().(type) {
	case *string:
		*v = val
	case *bool:
		*v = true
	}
}

func flagTakesValue(config reflect.Type, flag cmdtree.Flag) bool {
	// Everything except bool-typed fields take values
	return config.FieldByIndex(flag.Field).Type != reflect.TypeOf(true)
}
