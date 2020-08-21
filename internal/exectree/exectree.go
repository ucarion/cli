package exectree

import (
	"context"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/param"
)

func Exec(ctx context.Context, tree cmdtree.CommandTree, args []string) error {
	return exec(ctx, reflect.New(tree.Config).Elem(), tree, args)
}

func exec(ctx context.Context, config reflect.Value, tree cmdtree.CommandTree, args []string) error {
	posArgIndex := 0 // index of the next positional argument to assign to

	// Process args until there are no more args.
	for len(args) > 0 {
		var arg string // the arg we're processing next
		arg, args = args[0], args[1:]

		// What kind of argument are we dealing with?
		switch {
		case arg == "--":
			// It's the end-of-flags indicator. All remaining arguments are
			// strictly positional (or trailing); no subcommands, no flags.
			// We'll now drain all the remaining args into the positional and
			// trailing arguments, and that's it.
			for len(args) > 0 {
				arg, args = args[0], args[1:]

				if posArgIndex == len(tree.PosArgs) {
					// We have already used all the positional arguments. So
					// this argument must go into the trailing set of positional
					// arguments.

					// TODO handle no trailing posargs in tree
					if err := setConfigField(config, tree.TrailingArgs.Field, arg); err != nil {
						return err
					}
				} else {
					if err := setConfigField(config, tree.PosArgs[posArgIndex].Field, arg); err != nil {
						return err
					}

					posArgIndex++
				}
			}
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
				if err := setConfigField(config, flag.Field, value); err != nil {
					return err
				}
			} else {
				// The flag either doesn't take a value, or is in the separate
				// form, e.g. "--foo" or "--foo bar".
				flag, _ := getLongFlag(tree, arg[2:]) // TODO handle not there

				if mustTakeValue(tree.Config, flag) {
					// The flag takes a value. The next arg must be its value.
					arg, args = args[0], args[1:] // TODO no next arg
					if err := setConfigField(config, flag.Field, arg); err != nil {
						return err
					}
				} else {
					// The flag does not take a value. We just assign the
					// appropriate field to "true".
					if err := setConfigField(config, flag.Field, ""); err != nil {
						return err
					}
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
					// The flag does take a value. It may take on one of two
					// forms, which must be handled separately.
					if len(chars) > 0 {
						// The flag is in the "stuck" form, e.g. "-ojson". The flag's
						// value is the rest of the arg following the name of the flag.
						if err := setConfigField(config, flag.Field, chars); err != nil {
							return err
						}

						chars = "" // reset chars so we stop looking for more flags
					} else {
						// The flag is either in the "separate" form, e.g. "-o
						// json", or it doesn't *have* to take a value.
						if mustTakeValue(tree.Config, flag) {
							// The flag must take a value, so the next arg must
							// be the flag's value.
							arg, args = args[0], args[1:] // TODO no next arg
							if err := setConfigField(config, flag.Field, arg); err != nil {
								return err
							}
						} else {
							// The flag doesn't have to take a value. In such a
							// case, the "separate" form isn't applicable, and
							// the flag was merely "enabled", and not set to a
							// particular value.
							if err := setConfigField(config, flag.Field, ""); err != nil {
								return err
							}
						}
					}
				} else {
					// The flag does not take a value. We just assign the
					// appropriate field to "true".
					if err := setConfigField(config, flag.Field, ""); err != nil {
						return err
					}
				}
			}
		default:
			// This is either a positional argument or a subcommand. It's not
			// possible for a commmand to have both positional arguments and
			// sub-commands.
			for _, child := range tree.Children {
				if child.Name == arg {
					// This child is our argument. Forward along what we have
					// already to the child, and that's all we'll do here.
					childConfig := reflect.New(child.Config).Elem()
					childConfig.Field(child.ParentConfigField).Set(config)
					return exec(ctx, childConfig, child.CommandTree, args)
				}
			}

			// This is a positional argument. The next positional argument's
			// value is arg.
			if posArgIndex == len(tree.PosArgs) {
				// We have already used all the positional arguments. So this
				// argument must go into the trailing set of positional
				// arguments.

				// TODO handle no trailing posargs in tree
				if err := setConfigField(config, tree.TrailingArgs.Field, arg); err != nil {
					return err
				}
			} else {
				if err := setConfigField(config, tree.PosArgs[posArgIndex].Field, arg); err != nil {
					return err
				}

				posArgIndex++
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

func setConfigField(config reflect.Value, index []int, val string) error {
	// cmdtree.New will have handled making sure all fields are param-friendly.
	p, _ := param.New(config.FieldByIndex(index).Addr().Interface())
	return p.Set(val)
}

func mayTakeValue(config reflect.Type, flag cmdtree.Flag) bool {
	// Everything except bool-typed fields may take values
	return config.FieldByIndex(flag.Field).Type != reflect.TypeOf(true)
}

func mustTakeValue(config reflect.Type, flag cmdtree.Flag) bool {
	// Everything except bool-typed fields and point-typed fields take values
	return mayTakeValue(config, flag) && config.FieldByIndex(flag.Field).Type.Kind() != reflect.Ptr
}
