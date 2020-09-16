package cmdman

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/command"
	"github.com/ucarion/cli/internal/param"
)

func Man(tree cmdtree.CommandTree, name string) map[string]string {
	out := map[string]string{}
	walk(out, tree, []string{filepath.Base(name)})
	return out
}

func walk(out map[string]string, tree cmdtree.CommandTree, name []string) {
	k, v := man(tree, name)
	out[k] = v

	for childName, child := range tree.Children {
		walk(out, child.CommandTree, append(name, childName))
	}
}

func man(tree cmdtree.CommandTree, name []string) (string, string) {
	var buf bytes.Buffer

	// Initial header line.
	fmt.Fprintf(&buf, ".TH %s 1\n", strings.ToUpper(strings.Join(name, "-")))

	// Name section heading.
	fmt.Fprintln(&buf, ".SH NAME")

	// Either something like:
	//
	// foo - does foo things
	//
	// or
	//
	// foo
	//
	// Where "foo" is determined by the given name, and "does foo things" is the
	// short description of the command.
	if tree.Description != "" {
		fmt.Fprintf(&buf, "%s - %s\n", strings.Join(name, "-"), tree.Description)
	} else {
		fmt.Fprintln(&buf, strings.Join(name, "-"))
	}

	// Synopsis section. This should show how you can invoke the program.
	fmt.Fprintln(&buf, ".SH SYNOPSIS")
	fmt.Fprintf(&buf, "\\fI%s\\fR [<options>]", strings.Join(name, " "))

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
			fmt.Fprintf(&buf, " [%s]", strings.Join(children, " | "))
		} else {
			fmt.Fprintf(&buf, " %s", strings.Join(children, " | "))
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

	buf.WriteByte('\n')

	// Output the extended description. We'll do this unconditionally, because
	// the Description section is extremely common in man pages. We don't fall
	// back to the short description, because typographic conventions for that
	// message is different, and so is a poor fallback.
	fmt.Fprintln(&buf, ".SH DESCRIPTION")
	fmt.Fprintln(&buf, tree.ExtendedDescription)

	// Options section. This details each of the flags and their extended
	// usages.
	fmt.Fprintln(&buf, ".SH OPTIONS")
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
			flagLine = fmt.Sprintf("--%s%s", f.LongName, valuePart)
		case f.LongName == "":
			flagLine = fmt.Sprintf("-%s%s", f.ShortName, valuePart)
		default:
			flagLine = fmt.Sprintf("-%s, --%s%s", f.ShortName, f.LongName, valuePart)
		}

		fmt.Fprintln(&buf, ".TP")
		fmt.Fprintf(&buf, "%s\n", flagLine)
		fmt.Fprintln(&buf, f.ExtendedUsage)
	}

	// Return the name of the file this man page would go to, and its contents.
	//
	// We always want to generate a man page in the "1" section, because that is
	// where user commands go.
	return fmt.Sprintf("%s.1", strings.Join(name, "-")), buf.String()
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
