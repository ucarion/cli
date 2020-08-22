package cmdtree

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/param"
)

const TagSubCommand = "subcmd"
const TagCLI = "cli"

type CommandTree struct {
	Name         string
	Func         reflect.Value
	Config       reflect.Type
	Flags        []Flag
	PosArgs      []PosArg
	TrailingArgs PosArg
	Children     []ChildCommand
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
		// Ensure each of the inputted functions are valid.
		v := reflect.ValueOf(fn)
		if !isValidFunction(v) {
			// v.Type() will panic if fn is nil, so we return a separate error
			// in that case:
			if fn == nil {
				return CommandTree{}, fmt.Errorf("command funcs must be func(context.Context, T) error, got: %s", fn)
			}

			return CommandTree{}, fmt.Errorf("command funcs must be func(context.Context, T) error, got: %s", v.Type())
		}

		config, err := newConfigFromFunc(v)
		if err != nil {
			return CommandTree{}, err
		}

		configs = append(configs, config)
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
			config, err := newConfigFromType(t)
			if err != nil {
				return CommandTree{}, err
			}

			configs = append(configs, config)
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
	roots := children[nil]
	if len(roots) != 1 {
		return CommandTree{}, fmt.Errorf("multiple top-level commands")
	}

	return newCmd(children, roots[0]), nil
}

type config struct {
	ParentType reflect.Type
	ChildCommand
}

func newConfigFromFunc(fn reflect.Value) (config, error) {
	c, err := newConfigFromType(fn.Type().In(1))
	if err != nil {
		return config{}, err
	}

	c.CommandTree.Func = fn
	return c, nil
}

var errMultipleSubcmdTags = errors.New("multiple uses of subcmd tag in config struct")

func newConfigFromType(t reflect.Type) (config, error) {
	c := config{
		ChildCommand: ChildCommand{
			CommandTree: CommandTree{
				Config: t,
			},
		},
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if name, ok := f.Tag.Lookup(TagSubCommand); ok {
			// Ensure we haven't already seen a use of the tag already.
			if c.ParentType != nil {
				return config{}, errMultipleSubcmdTags
			}

			c.ParentType = f.Type
			c.ParentConfigField = i
			c.Name = name
		}

		if name, ok := getFieldOverrideName(f); ok {
			c.Name = name
		}
	}

	return c, addParamsFromType(&c, []int{}, t)
}

func addParamsFromType(c *config, indexPrefix []int, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		index := append(indexPrefix, i)

		if _, ok := getFieldOverrideName(f); ok {
			continue
		}

		if f.Anonymous {
			if err := addParamsFromType(c, index, f.Type); err != nil {
				return err
			}

			continue
		}

		cli, ok := f.Tag.Lookup(TagCLI)
		if !ok {
			continue
		}

		// This field has the cli tag on it. Is it a valid type to use a CLI
		// field?
		if err := checkValidFieldType(f.Type); err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}

		switch {
		case strings.HasPrefix(cli, "..."):
			if f.Type.Kind() != reflect.Slice {
				return fmt.Errorf("%s: trailing args must be a slice", f.Name)
			}

			c.TrailingArgs = PosArg{Field: index, Name: cli[3:]}
		case strings.HasPrefix(cli, "-"):
			longNames := []string{}
			shortNames := []string{}

			parts := strings.Split(cli, ",")
			for _, part := range parts {
				switch {
				case strings.HasPrefix(part, "--"):
					longNames = append(longNames, part[2:])
				case strings.HasPrefix(part, "-"):
					shortNames = append(shortNames, part[1:])
				}
			}

			c.Flags = append(c.Flags, Flag{
				Field:      index,
				LongNames:  longNames,
				ShortNames: shortNames,
			})

		default:
			c.PosArgs = append(c.PosArgs, PosArg{Field: index, Name: cli})
		}
	}

	return nil
}

func getFieldOverrideName(f reflect.StructField) (string, bool) {
	if f.Name == "_" && f.Type == reflect.StructOf(nil) {
		return f.Tag.Lookup(TagCLI)
	}

	return "", false
}

// This weird syntax is to get around the limitation that if you pass any sort
// of nil to reflect.TypeOf, you get back nil. The context and error types are
// used in checkValidFunction further below.
var (
	paramType = reflect.TypeOf((*param.Param)(nil)).Elem()
	ctxType   = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType   = reflect.TypeOf((*error)(nil)).Elem()
)

func checkValidFieldType(t reflect.Type) error {
	v := reflect.New(t).Interface()
	_, err := param.New(v)
	return err
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

func isValidFunction(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	t := v.Type()

	if v.Kind() != reflect.Func {
		return false
	}

	if t.NumIn() != 2 {
		return false
	}

	if !t.In(0).Implements(ctxType) {
		return false
	}

	if t.In(1).Kind() != reflect.Struct {
		return false
	}

	if t.NumOut() != 1 {
		return false
	}

	if !t.Out(0).Implements(errType) {
		return false
	}

	return true
}
