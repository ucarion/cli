package argparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/didyoumean"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/param"

	"github.com/ucarion/cli/internal/command"
)

type Parser struct {
	CommandTree     cmdtree.CommandTree
	Config          reflect.Value
	Name            []string
	ShowHelp        bool
	FlagsTerminated bool
	PosArgIndex     int
	Flag            command.Flag
	FlagIsShort     bool
}

func New(tree cmdtree.CommandTree) Parser {
	return Parser{
		CommandTree: tree,
		Config:      reflect.New(tree.Config).Elem(),
	}
}

func (p *Parser) ParseArg(s string) error {
	switch {
	case p.Name == nil:
		// If we don't have a name yet, then we must be parsing argv[0]. So we
		// shouldn't parse the string as an argument, but rather as the name of
		// our executable.
		p.Name = []string{s}

	case p.Flag.FieldIndex != nil:
		// We are currently in the state where the next arg is supposed to be
		// p.Flag's value. The given string is that value.
		if err := setConfigField(p.Config, p.Flag.FieldIndex, s); err != nil {
			if p.FlagIsShort {
				return fmt.Errorf("-%s: %w", p.Flag.ShortName, err)
			}

			return fmt.Errorf("--%s: %w", p.Flag.LongName, err)
		}

		p.Flag = command.Flag{}

	case p.FlagsTerminated:
		// We have previously reached the flag terminator argument ("--").
		// Regardless of any other context, we know that all remaining args
		// are positional.
		if err := p.parsePosArg(s); err != nil {
			return err
		}

	case s == "--":
		p.FlagsTerminated = true

	case strings.HasPrefix(s, "--") && strings.Contains(s, "="):
		// This is a long flag in the "stuck" form, e.g. "--foo=bar".
		//
		// It's valid to do "--foo=bar=baz", which sets "foo" to "bar=baz".
		// So we take care to only strip out the first "=" in arg. We also
		// strip out the leading dashes in "--foo" into "foo".
		parts := strings.SplitN(s, "=", 2)
		name, value := parts[0][2:], parts[1]

		flag, err := getLongFlag(p.CommandTree, name)
		if err != nil {
			return err
		}

		// The stuck form is illegal for flags that don't take a value. You
		// can't do "--foo=bar" if "--foo" doesn't take a value.
		if !mayTakeValue(p.Config, flag) {
			return fmt.Errorf("option --%s takes no value", name)
		}

		if err := setConfigField(p.Config, flag.FieldIndex, value); err != nil {
			return fmt.Errorf("--%s: %w", name, err)
		}

	case strings.HasPrefix(s, "--"):
		// This is a long flag in the "separate" form, e.g. "--foo bar".
		//
		// Here, we strip out the leading dashes in "--foo" into "foo".
		flag, err := getLongFlag(p.CommandTree, s[2:])
		if err != nil {
			return err
		}

		// If the flag doesn't have to take a value, then in the separate
		// form we just set its value to the empty string. This is the
		// documented contract for both boolean and optionally-taking-value
		// flags.
		if !mustTakeValue(p.Config, flag) {
			if flag.IsHelp {
				p.ShowHelp = true
				return nil
			}

			if err := setConfigField(p.Config, flag.FieldIndex, ""); err != nil {
				return fmt.Errorf("--%s: %w", s[2:], err)
			}

			return nil
		}

		// The value for flag will be in the next argument.
		p.Flag = flag
		p.FlagIsShort = false

	case strings.HasPrefix(s, "-"):
		// It's one or more short flags in a bundle. A "bundle" is a set of
		// flags like "-abc", which is an alias for "-a -b -c", assuming
		// "-a" and "-b" are boolean flags that don't take a value.
		//
		// Let's now consume the chars in arg one-by-one, to handle each
		// bundled flag. We skip the initial char, which is just a leading
		// dash.
		chars := s[1:]
		for len(chars) > 0 {
			var char string                           // the char we're processing next
			char, chars = string(chars[0]), chars[1:] // eat a char from chars

			flag, err := getShortFlag(p.CommandTree, char)
			if err != nil {
				return err
			}

			if mustTakeValue(p.Config, flag) && chars == "" {
				// Special-case the condition where the flag must take a value
				// and that value is in the next arg; this is the only case
				// where we'll need to update p.Flag.
				p.Flag = flag
				p.FlagIsShort = true
			} else if mayTakeValue(p.Config, flag) {
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
				if err := setConfigField(p.Config, flag.FieldIndex, chars); err != nil {
					return fmt.Errorf("-%s: %w", char, err)
				}

				chars = "" // stop looking through the bundle
			} else if flag.IsHelp {
				p.ShowHelp = true
			} else {
				// The flag doesn't take a value. Enable the flag, and keep
				// scanning the bundle.
				//
				// Setting a boolean flag can't fail.
				setConfigField(p.Config, flag.FieldIndex, "")
			}
		}

	default:
		// This argument is not a flag. It's either a positional argument or
		// a subcommand name. We don't need to worry about ambguity between
		// these cases; either Children is nonempty, or PosArgs/Trailing are
		// nonempty, but never both. This is enforced by cmdtree.New.
		if len(p.CommandTree.Children) != 0 {
			// We have children commands, so the arg must be a child command
			// name.
			child, ok := p.CommandTree.Children[s]
			if !ok {
				dym := didyoumean.DidYouMean(p.CommandTree, s)
				return fmt.Errorf("unknown sub-command: %s, did you mean: %s?", s, dym)
			}

			childConfig := reflect.New(child.Config).Elem()
			childConfig.Field(child.ParentIndexInChild).Set(p.Config)

			p.Config = childConfig
			p.CommandTree = child.CommandTree
			p.Name = append(p.Name, s)
		} else {
			// We don't have children commands, so the arg must be a positional
			// argument.
			if err := p.parsePosArg(s); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) parsePosArg(s string) error {
	var posArg command.PosArg
	if p.PosArgIndex == len(p.CommandTree.PosArgs) {
		posArg = p.CommandTree.Trailing
	} else {
		posArg = p.CommandTree.PosArgs[p.PosArgIndex]
		p.PosArgIndex++
	}

	if posArg.FieldIndex == nil {
		return fmt.Errorf("unexpected argument: %s", s)
	}

	if err := setConfigField(p.Config, posArg.FieldIndex, s); err != nil {
		return fmt.Errorf("%s: %w", posArg.Name, err)
	}

	return nil
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

func (p Parser) NoMoreArgs() error {
	if p.Flag.FieldIndex != nil {
		if p.FlagIsShort {
			return fmt.Errorf("option -%s requires a value", p.Flag.ShortName)
		} else {
			return fmt.Errorf("option --%s requires a value", p.Flag.LongName)
		}
	}

	if p.PosArgIndex < len(p.CommandTree.PosArgs) {
		posArg := p.CommandTree.PosArgs[p.PosArgIndex]
		return fmt.Errorf("argument %s requires a value", posArg.Name)
	}

	return nil
}
