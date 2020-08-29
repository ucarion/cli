package cmdhelp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/internal/command"
	"github.com/ucarion/cli/param"

	"github.com/ucarion/cli/internal/cmdtree"
)

func Help(tree cmdtree.CommandTree, name []string) string {
	posArgs := []string{}
	for _, a := range tree.PosArgs {
		posArgs = append(posArgs, a.Name)
	}

	if tree.Trailing.FieldIndex != nil {
		posArgs = append(posArgs, tree.Trailing.Name+"...")
	}

	posArgsPart := ""
	if len(posArgs) != 0 {
		posArgsPart = " " + strings.Join(posArgs, " ")
	}

	out := fmt.Sprintf("usage: %s [<options>]%s\n", strings.Join(name, " "), posArgsPart)

	if tree.ExtendedDescription != "" {
		out += "\n" + tree.ExtendedDescription + "\n"
	}

	out += "\n"

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
		out += fmt.Sprintf("    %-22s %s\n", flagLine, f.Usage)
	}

	return out
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
