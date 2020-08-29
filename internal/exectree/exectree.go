package exectree

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/cmdhelp"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/command"
	"github.com/ucarion/cli/param"
)

var HelpWriter io.Writer = os.Stdout

func Exec(ctx context.Context, tree cmdtree.CommandTree, args []string) error {
	return exec(ctx, reflect.New(tree.Config).Elem(), tree, args[:1], args[1:])
}

func exec(ctx context.Context, config reflect.Value, tree cmdtree.CommandTree, name []string, args []string) error {
	showHelp := false // whether an IsHelp flag was set
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
			var posArg command.PosArg
			if posArgIndex == len(tree.PosArgs) {
				posArg = tree.Trailing
			} else {
				posArg = tree.PosArgs[posArgIndex]
				posArgIndex++
			}

			if posArg.FieldIndex == nil {
				return fmt.Errorf("unexpected positional argument: %s", arg)
			}

			if err := setConfigField(config, posArg.FieldIndex, arg); err != nil {
				return fmt.Errorf("%s: %w", posArg.Name, err)
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

			if err := setConfigField(config, flag.FieldIndex, value); err != nil {
				return fmt.Errorf("--%s: %w", name, err)
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
				if flag.IsHelp {
					showHelp = true
					continue
				}

				if err := setConfigField(config, flag.FieldIndex, ""); err != nil {
					return fmt.Errorf("--%s: %w", arg[2:], err)
				}

				continue
			}

			// The value needs to be in the next arg. If there isn't a next arg,
			// that's an error.
			if len(args) == 0 {
				return fmt.Errorf("option --%s requires a value", arg[2:])
			}

			name := arg[2:]               // for error reporting
			arg, args = args[0], args[1:] // eat an arg from args
			if err := setConfigField(config, flag.FieldIndex, arg); err != nil {
				return fmt.Errorf("--%s: %w", name, err)
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
					if err := setConfigField(config, flag.FieldIndex, arg); err != nil {
						return fmt.Errorf("-%s: %w", char, err)
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
					if err := setConfigField(config, flag.FieldIndex, chars); err != nil {
						return fmt.Errorf("-%s: %w", char, err)
					}

					chars = "" // stop looking through the bundle
					continue
				}

				if flag.IsHelp {
					showHelp = true
					continue
				}

				// The flag doesn't take a value. Enable the flag, and keep
				// scanning the bundle.
				//
				// Setting a boolean flag can't fail.
				setConfigField(config, flag.FieldIndex, "")
			}

		default:
			// This argument is not a flag. It's either a positional argument or
			// a subcommand name. We don't need to worry about ambguity between
			// these cases; either Children is nonempty, or PosArgs/Trailing are
			// nonempty, but never both.
			//
			// Let's try to do the subcommand case first.
			if len(tree.Children) != 0 {
				child, ok := tree.Children[arg]
				if !ok {
					return fmt.Errorf("unknown sub-command: %s", arg)
				}

				childConfig := reflect.New(child.Config).Elem()
				childConfig.Field(child.ParentIndexInChild).Set(config)
				return exec(ctx, childConfig, child.CommandTree, append(name, arg), args)
			}

			// This code is the same as the positional argument logic in the
			// flagsTerminated branch.
			var posArg command.PosArg
			if posArgIndex == len(tree.PosArgs) {
				posArg = tree.Trailing
			} else {
				posArg = tree.PosArgs[posArgIndex]
				posArgIndex++
			}

			if posArg.FieldIndex == nil {
				return fmt.Errorf("unexpected positional argument: %s", arg)
			}

			if err := setConfigField(config, posArg.FieldIndex, arg); err != nil {
				return fmt.Errorf("%s: %w", posArg.Name, err)
			}
		}
	}

	if showHelp {
		_, err := HelpWriter.Write([]byte(cmdhelp.Help(tree, name)))
		return err
	}

	out := tree.Func.Call([]reflect.Value{reflect.ValueOf(ctx), config})

	err := out[0].Interface()
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", strings.Join(name, " "), err.(error))
}

func getLongFlag(tree cmdtree.CommandTree, s string) (command.Flag, error) {
	for _, f := range tree.Flags {
		if f.LongName == s {
			return f, nil
		}
	}

	return command.Flag{}, fmt.Errorf("unknown option: --%s", s)
}

func getShortFlag(tree cmdtree.CommandTree, s string) (command.Flag, error) {
	for _, f := range tree.Flags {
		if f.ShortName == s {
			return f, nil
		}
	}

	return command.Flag{}, fmt.Errorf("unknown option: -%s", s)
}

func setConfigField(config reflect.Value, index []int, val string) error {
	// cmdtree.New will have handled making sure all fields are param-friendly.
	p, _ := param.New(config.FieldByIndex(index).Addr().Interface())
	return p.Set(val)
}

func mayTakeValue(config reflect.Value, flag command.Flag) bool {
	if flag.IsHelp {
		return false
	}

	p, _ := param.New(config.FieldByIndex(flag.FieldIndex).Addr().Interface())
	return param.MayTakeValue(p)
}

func mustTakeValue(config reflect.Value, flag command.Flag) bool {
	if flag.IsHelp {
		return false
	}

	p, _ := param.New(config.FieldByIndex(flag.FieldIndex).Addr().Interface())
	return param.MustTakeValue(p)
}
