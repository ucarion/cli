package cmdtree

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ucarion/cli/value"
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
		if err := checkValidFunction(v); err != nil {
			return CommandTree{}, err
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
	root := newCmd(children, children[nil][0]) // TODO assert only one root

	return root, nil
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

var ErrMultipleSubcmdTags = errors.New("multiple uses of subcmd tag in struct")

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

		// TODO assert only one use of the tag in the struct
		if name, ok := f.Tag.Lookup(TagSubCommand); ok {
			// Ensure we haven't already seen a use of the tag already.
			if c.ParentType != nil {
				return config{}, ErrMultipleSubcmdTags
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

type InvalidConfigFieldTypeErr struct {
	Type reflect.Type
}

func (e InvalidConfigFieldTypeErr) Error() string {
	return fmt.Sprintf("bad config field type: %v", e.Type)
}

var ErrTrailingMustBeSlice = errors.New("trailing args field must be a slice")

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
			return err
		}

		switch {
		case strings.HasPrefix(cli, "..."):
			if f.Type.Kind() != reflect.Slice {
				return ErrTrailingMustBeSlice
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
	valueType = reflect.TypeOf((*value.Value)(nil)).Elem()
	ctxType   = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType   = reflect.TypeOf((*error)(nil)).Elem()
)

func checkValidFieldType(t reflect.Type) error {
	err := InvalidConfigFieldTypeErr{Type: t}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()

		// Pointers-of-pointers or pointers-of-bool are invalid field types.
		// Pointers in structs indicate optionally-taking-value types, but it
		// doesn't make sense to have "twice-optionally" fields, nor does it
		// make sense to have optionally-taking-value bool fields, since bool
		// fields never take a value.
		if t.Kind() == reflect.Bool || t.Kind() == reflect.Ptr {
			return err
		}
	}

	// Slices are permitted only if they contain something that's a valid type.
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	switch t.Kind() {
	// All of the primitive types are supported directly.
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return nil
	case reflect.Struct:
		// A struct is valid only if it implements Value.
		if reflect.PtrTo(t).Implements(valueType) {
			return nil
		}

		return err
	default:
		return err
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

type NotValidFunctionErr struct {
	Value reflect.Value
}

func (e NotValidFunctionErr) Error() string {
	return fmt.Sprintf("bad func - want func(context.Context, T) error, got: %v", e.Value.Type())
}

func checkValidFunction(v reflect.Value) error {
	err := NotValidFunctionErr{v}
	if !v.IsValid() {
		return err
	}

	t := v.Type()

	if v.Kind() != reflect.Func {
		return err
	}

	if t.NumIn() != 2 {
		return err
	}

	if !t.In(0).Implements(ctxType) {
		return err
	}

	if t.In(1).Kind() != reflect.Struct {
		return err
	}

	if t.NumOut() != 1 {
		return err
	}

	if !t.Out(0).Implements(errType) {
		return err
	}

	return nil
}
