package exectree

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/param"
)

func Exec(ctx context.Context, tree cmdtree.CommandTree, args []string) error {
	return exec(ctx, reflect.New(tree.Config).Elem(), tree, args)
}

func exec(ctx context.Context, config reflect.Value, tree cmdtree.CommandTree, args []string) error {
	flagsTerminated := false
	posArgIndex := 0 // index of the next positional argument to assign to

	// Process args until there are no more args.
	for len(args) > 0 {
		var arg string                // the arg we're processing next
		arg, args = args[0], args[1:] // eat one arg from args

		switch {
		case flagsTerminated:
			// We have previously reached the flag terminator argument ("--").
			// Regardless of any other context, we know that all remaining args
			// are positional.
			var posArg cmdtree.PosArg
			if posArgIndex == len(tree.PosArgs) {
				posArg = tree.TrailingArgs
			} else {
				posArg = tree.PosArgs[posArgIndex]
				posArgIndex++
			}

			if posArg.Field == nil {
				return fmt.Errorf("unexpected positional argument: %s", arg)
			}

			if err := setConfigField(config, posArg.Field, arg); err != nil {
				return err
			}

		case arg == "--":
			// This is the flag terminator. We ignore this arg itself, but will
			// not look for further positional arguments or subcommands going
			// forward.
			flagsTerminated = true

		case strings.HasPrefix(arg, "--") && strings.Contains(arg, "="):
			// This is a long flag in the "stuck" form, e.g. "--foo=bar".
			//
			// It's valid to do "--foo=bar=baz", which sets "foo" to "bar=baz".
			// So we take care to only strip out the first "=" in arg. We also
			// strip out the leading dashes in "--foo" into "foo".
			parts := strings.SplitN(arg, "=", 2)
			name, value := parts[0][2:], parts[1]

			flag, err := getLongFlag(tree, name)
			if err != nil {
				return err
			}

			// The stuck form is illegal for flags that don't take a value. You
			// can't do "--foo=bar" if "--foo" doesn't take a value.
			if !mayTakeValue(config, flag) {
				return fmt.Errorf("option --%s takes no value", name)
			}

			if err := setConfigField(config, flag.Field, value); err != nil {
				return err
			}

		case strings.HasPrefix(arg, "--"):
			// This is a long flag in the "separate" form, e.g. "--foo bar".
			//
			// Here, we strip out the leading dashes in "--foo" into "foo".
			flag, err := getLongFlag(tree, arg[2:])
			if err != nil {
				return err
			}

			// If the flag doesn't have to take a value, then in the separate
			// form we just set its value to the empty string. This is the
			// documented contract for both boolean and optionally-taking-value
			// flags.
			if !mustTakeValue(config, flag) {
				if err := setConfigField(config, flag.Field, ""); err != nil {
					return err
				}

				continue
			}

			// The value needs to be in the next arg. If there isn't a next arg,
			// that's an error.
			if len(args) == 0 {
				return fmt.Errorf("option --%s requires a value", arg[2:])
			}

			arg, args = args[0], args[1:] // eat an arg from args
			if err := setConfigField(config, flag.Field, arg); err != nil {
				return err
			}

		case strings.HasPrefix(arg, "-"):
			// It's one or more short flags in a bundle. A "bundle" is a set of
			// flags like "-abc", which is an alias for "-a -b -c", assuming
			// "-a" and "-b" are boolean flags that don't take a value.
			//
			// Let's now consume the chars in arg one-by-one, to handle each
			// bundled flag. We skip the initial char, which is just a leading
			// dash.
			chars := arg[1:]
			for len(chars) > 0 {
				var char string                           // the char we're processing next
				char, chars = string(chars[0]), chars[1:] // eat a char from chars

				flag, err := getShortFlag(tree, char)
				if err != nil {
					return err
				}

				// Special-case the condition where the flag must take a value
				// and that value is in the next arg; this is the only case
				// where we'll need to eat an arg from args.
				if mustTakeValue(config, flag) && chars == "" {
					// There is no data left in the arg. The next arg must be
					// its value. If there isn't a next arg, that's an error.
					if len(args) == 0 {
						return fmt.Errorf("option -%s requires a value", char)
					}

					arg, args = args[0], args[1:] // eat an arg from args
					if err := setConfigField(config, flag.Field, arg); err != nil {
						return err
					}

					chars = "" // stop looking through the bundle
					continue
				}

				// If the flag can may take a value, then the rest of the arg is
				// its value.
				//
				// Slightly subtle code here: even if chars is empty, this code
				// is still correct; if chars is empty, then this branch of code
				// only runs for optionally-taking-value flags (else the
				// if-block above would have done the job).
				//
				// The contract for optionally-taking-value flags is that we set
				// them to empty-string if the value wasn't provided; that's
				// precisely what we'll do if chars is empty in this if-block.
				if mayTakeValue(config, flag) {
					if err := setConfigField(config, flag.Field, chars); err != nil {
						return err
					}

					chars = "" // stop looking through the bundle
					continue
				}

				// The flag doesn't take a value. Enable the flag, and keep
				// scanning the bundle.
				//
				// Setting a boolean flag can't fail.
				setConfigField(config, flag.Field, "")
			}

		default:
			// This argument is not a flag. It's either a positional argument or
			// a subcommand name. We don't need to worry about ambguity between
			// these cases; either Children is nonempty, or PosArgs/Trailing are
			// nonempty, but never both.
			//
			// Let's try to do the subcommand case first.
			for _, child := range tree.Children {
				if child.Name == arg {
					// This child is our argument. Forward along what we have
					// already to the child, and let the child process the
					// remaining args.
					childConfig := reflect.New(child.Config).Elem()
					childConfig.Field(child.ParentConfigField).Set(config)
					return exec(ctx, childConfig, child.CommandTree, args)
				}
			}

			// This code is the same as the positional argument logic in the
			// flagsTerminated branch.
			var posArg cmdtree.PosArg
			if posArgIndex == len(tree.PosArgs) {
				posArg = tree.TrailingArgs
			} else {
				posArg = tree.PosArgs[posArgIndex]
				posArgIndex++
			}

			if posArg.Field == nil {
				return fmt.Errorf("unexpected positional argument: %s", arg)
			}

			if err := setConfigField(config, posArg.Field, arg); err != nil {
				return err
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

func getLongFlag(tree cmdtree.CommandTree, s string) (cmdtree.Flag, error) {
	for _, f := range tree.Flags {
		for _, name := range f.LongNames {
			if name == s {
				return f, nil
			}
		}
	}

	return cmdtree.Flag{}, fmt.Errorf("unknown option: --%s", s)
}

func getShortFlag(tree cmdtree.CommandTree, s string) (cmdtree.Flag, error) {
	for _, f := range tree.Flags {
		for _, name := range f.ShortNames {
			if name == s {
				return f, nil
			}
		}
	}

	return cmdtree.Flag{}, fmt.Errorf("unknown option: -%s", s)
}

func setConfigField(config reflect.Value, index []int, val string) error {
	// cmdtree.New will have handled making sure all fields are param-friendly.
	p, _ := param.New(config.FieldByIndex(index).Addr().Interface())
	return p.Set(val)
}

func mayTakeValue(config reflect.Value, flag cmdtree.Flag) bool {
	p, _ := param.New(config.FieldByIndex(flag.Field).Addr().Interface())
	return param.MayTakeValue(p)
}

func mustTakeValue(config reflect.Value, flag cmdtree.Flag) bool {
	p, _ := param.New(config.FieldByIndex(flag.Field).Addr().Interface())
	return param.MustTakeValue(p)
}
