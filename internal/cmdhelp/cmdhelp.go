package cmdhelp

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/command"
	"github.com/ucarion/cli/internal/param"
)

func Help(tree cmdtree.CommandTree, name []string) string {
	var buf bytes.Buffer

	// First, write the beginning of the usage line.
	fmt.Fprintf(&buf, "usage: %s [<options>]", strings.Join(name, " "))

	// Next, we'll write either the valid sub-commands or the valid positional
	// arguments of the command.
	if tree.Children != nil {
		// The command has sub-commands, so we'll output those.
		children := []string{}
		for k := range tree.Children {
			children = append(children, k)
		}

		sort.Strings(children)

		// If the tree is itself executable, then sub-commands are optional and
		// so are wrapped in square brackets.
		if tree.Func.IsValid() {
			fmt.Fprintf(&buf, " [%s]", strings.Join(children, "|"))
		} else {
			fmt.Fprintf(&buf, " %s", strings.Join(children, "|"))
		}
	} else {
		// The command doesn't have sub-commands, so we'll output positional
		// args, if there are any.
		posArgs := []string{}
		for _, a := range tree.PosArgs {
			posArgs = append(posArgs, a.Name)
		}

		if tree.Trailing.FieldIndex != nil {
			posArgs = append(posArgs, tree.Trailing.Name+"...")
		}

		if len(posArgs) != 0 {
			fmt.Fprintf(&buf, " %s", strings.Join(posArgs, " "))
		}
	}

	// Finish the initial usage line.
	buf.WriteByte('\n')

	// If there's an extended description, write it out with surrounding
	// newlines.
	if tree.ExtendedDescription != "" {
		fmt.Fprintf(&buf, "\n%s\n", tree.ExtendedDescription)
	}

	// Insert a blank line before the flags.
	buf.WriteByte('\n')

	// Write out the flags.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	for _, f := range tree.Flags {
		valueName := f.ValueName
		if valueName == "" {
			valueName = flagValueType(tree, f).String()
		}

		var valuePart string
		switch {
		case mustTakeValue(tree, f):
			valuePart = fmt.Sprintf(" <%s>", valueName)
		case mayTakeValue(tree, f):
			valuePart = fmt.Sprintf("[=<%s>]", valueName)
		}

		var flagLine string
		switch {
		case f.ShortName == "":
			flagLine = fmt.Sprintf("    --%s%s", f.LongName, valuePart)
		case f.LongName == "":
			flagLine = fmt.Sprintf("-%s%s", f.ShortName, valuePart)
		default:
			flagLine = fmt.Sprintf("-%s, --%s%s", f.ShortName, f.LongName, valuePart)
		}

		fmt.Fprintf(w, "    %s\t   %s\n", flagLine, f.Usage)
	}

	w.Flush()

	// Add one last empty line to make the output more clearly separated from
	// the subsequent CLI prompt.
	buf.WriteByte('\n')

	return buf.String()
}

func flagValueType(tree cmdtree.CommandTree, flag command.Flag) reflect.Type {
	return tree.Config.FieldByIndex(flag.FieldIndex).Type
}

func mayTakeValue(tree cmdtree.CommandTree, flag command.Flag) bool {
	if flag.IsHelp {
		return false
	}

	config := reflect.New(tree.Config).Elem()
	p, _ := param.New(config.FieldByIndex(flag.FieldIndex).Addr().Interface())
	return param.MayTakeValue(p)
}

func mustTakeValue(tree cmdtree.CommandTree, flag command.Flag) bool {
	if flag.IsHelp {
		return false
	}

	config := reflect.New(tree.Config).Elem()
	p, _ := param.New(config.FieldByIndex(flag.FieldIndex).Addr().Interface())
	return param.MustTakeValue(p)
}
