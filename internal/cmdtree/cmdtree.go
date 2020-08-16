package cmdtree

import "reflect"

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
	Child             CommandTree
}

func New(funcs []interface{}) (CommandTree, error) {
	f := reflect.ValueOf(funcs[0])

	return CommandTree{
		Config: f.Type().In(1),
		Func:   f,
	}, nil
}
