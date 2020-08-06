package cmdtree

import (
	"errors"
	"reflect"

	"github.com/ucarion/cli/internal/cmd"
)

type CommandTree struct {
	Root     cmd.Command
	Children map[string]CommandTree
}

var ErrMultipleRoots = errors.New("cmdtree: multiple roots")

func FromFuncs(fns []interface{}) (CommandTree, error) {
	// Nodes in the command tree
	cmds := []cmd.Command{}

	// Map from parent types to elements in cmds that have it as parent
	parentTypes := map[reflect.Type][]int{}

	// Types that we have not yet added to cmds
	toAdd := []reflect.Type{}

	// First, construct the executable nodes in the command tree
	for _, fn := range fns {
		c, parentType, err := cmd.FromFunc(fn)
		if err != nil {
			return CommandTree{}, err
		}

		cmds = append(cmds, c)

		if _, ok := parentTypes[parentType]; !ok {
			parentTypes[parentType] = []int{}
			if parentType != nil {
				toAdd = append(toAdd, parentType)
			}
		}

		parentTypes[parentType] = append(parentTypes[parentType], len(cmds)-1)
	}

	// Next, construct intermediate configs from the known parent types. Keep
	// doing this until there are no new types to explore.
	lastToAddLen := 0
	for lastToAddLen < len(toAdd) {
		lastToAddLen = len(toAdd)

		for _, t := range toAdd {
			c, parentType, err := cmd.FromConfig(t)
			if err != nil {
				return CommandTree{}, err
			}

			cmds = append(cmds, c)

			if _, ok := parentTypes[parentType]; !ok {
				parentTypes[parentType] = []int{}
				if parentType != nil {
					toAdd = append(toAdd, parentType)
				}
			}

			parentTypes[parentType] = append(parentTypes[parentType], len(cmds)-1)
		}
	}

	roots := toTree(cmds, parentTypes, nil)
	if len(roots) > 1 {
		return CommandTree{}, ErrMultipleRoots
	}

	return roots[0], nil
}

func toTree(cmds []cmd.Command, parentTypes map[reflect.Type][]int, rootType reflect.Type) []CommandTree {
	out := []CommandTree{}
	for _, i := range parentTypes[rootType] {
		children := map[string]CommandTree{}
		for _, child := range toTree(cmds, parentTypes, cmds[i].ConfigType) {
			children[child.Root.Name] = child
		}

		out = append(out, CommandTree{
			Root:     cmds[i],
			Children: children,
		})
	}

	return out
}
