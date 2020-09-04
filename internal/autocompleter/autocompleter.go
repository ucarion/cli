package autocompleter

import (
	"reflect"
	"sort"

	"github.com/ucarion/cli/internal/argparser"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/command"
)

func Autocomplete(tree cmdtree.CommandTree, args []string) []string {
	parser := argparser.New(tree)

	for _, arg := range args {
		if err := parser.ParseArg(arg); err != nil {
			// An autocompleter can't usefully return an error. If we can't
			// parse args successfully, we just won't return any suggestions.
			return nil
		}
	}

	if parser.Flag.FieldIndex != nil {
		// We are expecting a flag's value next. If that flag has an
		// autocomplete func, we'll return that func's results. Otherwise, we
		// have no suggestions.
		if !parser.Flag.AutocompleteFunc.IsValid() {
			return nil
		}

		out := parser.Flag.AutocompleteFunc.Call([]reflect.Value{parser.Config})
		return out[0].Interface().([]string)
	}

	var out []string

	// As long as flags aren't terminated, then the next argument could be a
	// flag.
	if !parser.FlagsTerminated {
		for _, f := range parser.CommandTree.Flags {
			// Don't include help flags.
			if f.IsHelp {
				continue
			}

			// If the config value for this flag is non-zero, then we assume
			// that flag has been used, and so we do not include it in the
			// autocompletion suggestions.
			if !parser.Config.FieldByIndex(f.FieldIndex).IsZero() {
				continue
			}

			if f.LongName != "" {
				out = append(out, "--"+f.LongName)
			} else {
				out = append(out, "-"+f.ShortName)
			}
		}
	}

	// If the flag has children commands, then suggest those children command
	// names.
	if parser.CommandTree.Children != nil {
		for childCmd := range parser.CommandTree.Children {
			out = append(out, childCmd)
		}

		// The rest of the possible suggestions are for posargs, which we cannot
		// have because we instead have child commands.
		sort.Strings(out)
		return out
	}

	var posArg command.PosArg
	if parser.PosArgIndex == len(parser.CommandTree.PosArgs) {
		posArg = parser.CommandTree.Trailing
	} else if parser.CommandTree.PosArgs != nil {
		posArg = parser.CommandTree.PosArgs[parser.PosArgIndex]
	}

	if posArg.FieldIndex != nil {
		if posArg.AutocompleteFunc.IsValid() {
			fnOut := posArg.AutocompleteFunc.Call([]reflect.Value{parser.Config})
			out = append(out, fnOut[0].Interface().([]string)...)
		}
	}

	sort.Strings(out)
	return out
}
