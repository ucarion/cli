package cmdtree

import (
	"fmt"
	"reflect"

	"github.com/ucarion/cli/internal/command"
)

type CommandTree struct {
	Children map[string]ChildCommand
	command.Command
}

type ChildCommand struct {
	ParentIndexInChild int
	CommandTree
}

type cmdWithParentInfo struct {
	command.Command
	command.ParentInfo
}

func New(fns []interface{}) (CommandTree, error) {
	cmds := []cmdWithParentInfo{}
	for _, fn := range fns {
		cmd, pinfo, err := command.FromFunc(fn)
		if err != nil {
			return CommandTree{}, err
		}

		cmds = append(cmds, cmdWithParentInfo{Command: cmd, ParentInfo: pinfo})
	}

	for {
		typesToAdd := map[reflect.Type]struct{}{}

		for _, cmd := range cmds {
			if cmd.ParentType == nil {
				continue
			}

			// Does this command have its parent type in cmds?
			ok := false
			for _, c := range cmds {
				if c.Config == cmd.ParentType {
					ok = true
				}
			}

			if !ok {
				typesToAdd[cmd.ParentType] = struct{}{}
			}
		}

		for t := range typesToAdd {
			cmd, pinfo, err := command.FromType(t)
			if err != nil {
				return CommandTree{}, err
			}

			cmds = append(cmds, cmdWithParentInfo{Command: cmd, ParentInfo: pinfo})
		}

		if len(typesToAdd) == 0 {
			break
		}
	}

	cmdsByParent := map[reflect.Type][]cmdWithParentInfo{}
	for _, cmd := range cmds {
		cmdsByParent[cmd.ParentType] = append(cmdsByParent[cmd.ParentType], cmd)
	}

	roots := cmdsByParent[nil]
	if len(roots) != 1 {
		rootTypes := []reflect.Type{}
		for _, root := range roots {
			rootTypes = append(rootTypes, root.Config)
		}

		return CommandTree{}, fmt.Errorf("multiple top-level commands: %v", rootTypes)
	}

	return CommandTree{
		Command:  roots[0].Command,
		Children: newForest(cmdsByParent, roots[0].Config),
	}, nil
}

func newForest(cmdsByParent map[reflect.Type][]cmdWithParentInfo, root reflect.Type) map[string]ChildCommand {
	out := map[string]ChildCommand{}
	for _, cmd := range cmdsByParent[root] {
		out[cmd.ChildName] = ChildCommand{
			ParentIndexInChild: cmd.ParentIndexInChild,
			CommandTree: CommandTree{
				Command:  cmd.Command,
				Children: newForest(cmdsByParent, cmd.Config),
			},
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
