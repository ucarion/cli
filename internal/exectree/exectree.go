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
		case strings.HasPrefix(arg, "--"):
			// It's a long flag. It may be either in the stuck form or the
			// separate form. We can distinguish these cases by seeing if this
			// arg has an equals sign or not.
			if strings.Contains(arg, "=") {
				// The flag is in the stuck form, e.g. "--foo=bar".
				//
				// It's valid to do "--foo=bar=baz", in which case the flag is
				// "foo" and the value is "bar=baz". So we only split off the
				// first "=".
				argParts := strings.SplitN(arg, "=", 2)
				name, value := argParts[0][2:], argParts[1]

				flag, _ := getLongFlag(tree, name) // TODO handle not there
				setConfigField(config, flag.Field, value)
			} else {
				// The flag either doesn't take a value, or is in the separate
				// form, e.g. "--foo" or "--foo bar".
				flag, _ := getLongFlag(tree, arg[2:]) // TODO handle not there

				if mustTakeValue(tree.Config, flag) {
					// The flag takes a value. The next arg must be its value.
					arg, args = args[0], args[1:] // TODO no next arg
					setConfigField(config, flag.Field, arg)
				} else {
					// The flag does not take a value. We just assign the
					// appropriate field to "true".
					setConfigField(config, flag.Field, "")
				}
			}
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

				// Can the flag take a value?
				if mayTakeValue(tree.Config, flag) {
					fmt.Println("chars", flag, chars)
					// The flag does take a value. It may take on one of two
					// forms, which must be handled separately.
					if len(chars) > 0 {
						// The flag is in the "stuck" form, e.g. "-ojson". The flag's
						// value is the rest of the arg following the name of the flag.
						setConfigField(config, flag.Field, chars)
						chars = "" // reset chars so we stop looking for more flags
					} else {
						// The flag is either in the "separate" form, e.g. "-o
						// json", or it doesn't *have* to take a value.
						if mustTakeValue(tree.Config, flag) {
							// The flag must take a value, so the next arg must
							// be the flag's value.
							arg, args = args[0], args[1:] // TODO no next arg
							setConfigField(config, flag.Field, arg)
						} else {
							// The flag doesn't have to take a value. In such a
							// case, the "separate" form isn't applicable, and
							// the flag was merely "enabled", and not set to a
							// particular value.
							setConfigField(config, flag.Field, "")
						}
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

func getLongFlag(tree cmdtree.CommandTree, s string) (cmdtree.Flag, bool) {
	for _, f := range tree.Flags {
		for _, name := range f.LongNames {
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
	case **string:
		*v = &val
	case *string:
		*v = val
	case *bool:
		*v = true
	}
}

func mayTakeValue(config reflect.Type, flag cmdtree.Flag) bool {
	// Everything except bool-typed fields may take values
	return config.FieldByIndex(flag.Field).Type != reflect.TypeOf(true)
}

func mustTakeValue(config reflect.Type, flag cmdtree.Flag) bool {
	// Everything except bool-typed fields and point-typed fields take values
	return mayTakeValue(config, flag) && config.FieldByIndex(flag.Field).Type.Kind() != reflect.Ptr
}
