package cmdtree

import (
	"reflect"
	"strings"
)

const TagSubCommand = "subcmd"
const TagCLI = "cli"

type CommandTree struct {
	Name     string
	Func     reflect.Value
	Config   reflect.Type
	Flags    []Flag
	PosArgs  []PosArg
	Children []ChildCommand
}

type Flag struct {
	Field      []int
	ShortNames []string
	LongNames  []string
}

type PosArg struct {
	Field []int
	Name  string
}

type ChildCommand struct {
	ParentConfigField int
	CommandTree
}

func New(funcs []interface{}) (CommandTree, error) {
	// First, we need to discover all nodes in the graph. We'll construct an
	// index from config types (which must be unique per func) to the one or
	// zero funcs that use the config type.
	//
	// Next, we'll construct a map of config types to their parent config types.
	// Each config type has exactly one or zero parent config types.
	//
	// From such a map, we'll "reverse" the direction of the relationship to
	// construct a forest.
	//
	// One tricky aspect of this is keeping track of ParentConfigField. It's not
	// enough to know what the parent config is; we must also keep track of what
	// index the parent config is in.

	// Construct the initial set of configs.
	configs := []config{}
	for _, fn := range funcs {
		// TODO verify fn is func(context.Context, T) error
		configs = append(configs, newConfigFromFunc(reflect.ValueOf(fn)))
	}

	// For each config, see if it has a parent type we don't already have in the
	// set of configs. If so, add that parent type to the set of configs. Do
	// this until a pass does not uncover any new configs.
	newConfigsAdded := true
	for newConfigsAdded {
		// Multiple configs may want to add the same parent config type, so
		// we'll dedupe them with a map.
		typesToAdd := map[reflect.Type]struct{}{}
		for _, config := range configs {
			// Does the config have a parent type at all?
			if config.ParentType == nil {
				continue
			}

			// See if there's an existing config whose type is our ParentType.
			ok := false
			for _, c := range configs {
				if c.CommandTree.Config == config.ParentType {
					ok = true
				}
			}

			// Our ParentType is not accounted for. Add it to the set of types
			// to explore.
			if !ok {
				typesToAdd[config.ParentType] = struct{}{}
			}
		}

		// For each type to add, construct an instance of config and add it to
		// the known configs.
		for t := range typesToAdd {
			configs = append(configs, newConfigFromType(t))
		}

		// If we had any types to add, then make sure we do another pass.
		newConfigsAdded = len(typesToAdd) > 0
	}

	// With all the edges of our command graph constructed, let's now reverse
	// the direction of the edges. We'll then start from the "nil" parent type
	// to construct our tree.
	children := map[reflect.Type][]config{} // map from parent types to configs
	for _, c := range configs {
		children[c.ParentType] = append(children[c.ParentType], c)
	}

	// We should expect to have exactly one child of the parent type "nil", and
	// that child is the root command.
	root := newCmd(children, children[nil][0]) // TODO assert only one root

	return root, nil
}

type config struct {
	ParentType reflect.Type
	ChildCommand
}

func newConfigFromFunc(fn reflect.Value) config {
	c := newConfigFromType(fn.Type().In(1))
	c.CommandTree.Func = fn
	return c
}

func newConfigFromType(t reflect.Type) config {
	c := config{
		ChildCommand: ChildCommand{
			CommandTree: CommandTree{
				Config: t,
			},
		},
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// TODO assert only one use of the tag in the struct
		if name, ok := f.Tag.Lookup(TagSubCommand); ok {
			c.ParentType = f.Type
			c.ParentConfigField = i
			c.Name = name
		}

		// Special-case for manually naming a command: a field named "_" whose
		// type is an empty struct, and which has a "cli" tag".
		if f.Name == "_" && f.Type == reflect.StructOf(nil) {
			if name, ok := f.Tag.Lookup(TagCLI); ok {
				c.Name = name
			}
		}
	}

	addParamsFromType(&c, []int{}, t)

	return c
}

func addParamsFromType(c *config, indexPrefix []int, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		index := append(indexPrefix, i)

		if f.Anonymous {
			addParamsFromType(c, index, f.Type)
			continue
		}

		name, ok := f.Tag.Lookup(TagCLI)
		if !ok {
			continue
		}

		longNames := []string{}
		shortNames := []string{}

		parts := strings.Split(name, ",")
		for _, part := range parts {
			switch {
			case strings.HasPrefix(part, "--"):
				longNames = append(longNames, part[2:])
			case strings.HasPrefix(part, "-"):
				shortNames = append(shortNames, part[1:])
			}
		}

		if len(longNames) > 0 || len(shortNames) > 0 {
			c.Flags = append(c.Flags, Flag{
				Field:      index,
				LongNames:  longNames,
				ShortNames: shortNames,
			})
		}
	}
}

func newCmd(children map[reflect.Type][]config, root config) CommandTree {
	root.Children = newChildCmds(children, root.CommandTree.Config)
	return root.CommandTree
}

func newChildCmds(children map[reflect.Type][]config, root reflect.Type) []ChildCommand {
	out := []ChildCommand{}
	for _, c := range children[root] {
		c.CommandTree = newCmd(children, c)
		out = append(out, c.ChildCommand)
	}

	return out
}
