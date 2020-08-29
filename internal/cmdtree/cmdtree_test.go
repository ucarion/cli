package cmdtree_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/command"
)

var helpFlag = command.Flag{
	IsHelp:        true,
	ShortName:     "h",
	LongName:      "help",
	Usage:         "display this help and exit",
	ExtendedUsage: "Display help message and exit.",
}

func TestNew_Basic(t *testing.T) {
	type args struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Command: command.Command{
			Config: reflect.TypeOf(args{}),
			Flags:  []command.Flag{helpFlag},
		},
	}, removeFunc(tree))
}

func TestNew_BadCommand(t *testing.T) {
	_, err := cmdtree.New([]interface{}{
		func() {},
	})

	assert.Equal(t,
		"command funcs must have type: func(context.Context, T) error, got: func()",
		err.Error())
}

func TestNew_BadConfigType(t *testing.T) {
	type rootArgs struct {
		X string `cli:"_"`
	}

	type args struct {
		Root rootArgs `cli:"root,subcmd"`
	}

	_, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.Equal(t,
		"invalid positional argument name: _",
		err.Error())
}

func TestNew_RootAndSub(t *testing.T) {
	type root struct{}
	type sub struct {
		Root root `cli:"sub,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ root) error { return nil },
		func(_ context.Context, _ sub) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Command: command.Command{
			Config: reflect.TypeOf(root{}),
			Flags:  []command.Flag{helpFlag},
		},
		Children: map[string]cmdtree.ChildCommand{
			"sub": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(sub{}),
						Flags:  []command.Flag{helpFlag},
					},
				},
			},
		},
	}, removeFunc(tree))
}

func TestNew_TwoSubcommands(t *testing.T) {
	type root struct{}

	type foo struct {
		Root root `cli:"foo,subcmd"`
	}

	type bar struct {
		Root root `cli:"bar,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ foo) error { return nil },
		func(_ context.Context, _ bar) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Command: command.Command{
			Config: reflect.TypeOf(root{}),
			Flags:  []command.Flag{helpFlag},
		},
		Children: map[string]cmdtree.ChildCommand{
			"foo": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(foo{}),
						Flags:  []command.Flag{helpFlag},
					},
				},
			},
			"bar": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(bar{}),
						Flags:  []command.Flag{helpFlag},
					},
				},
			},
		},
	}, removeFunc(tree))
}

func TestNew_ComplexTree(t *testing.T) {
	// Tests the following tree:
	//
	// 	root
	// 		a (runnable)
	// 		b (runnable)
	// 			c (runnable)
	// 			e
	// 				f (runnable)
	// 		g
	// 			h (runnable)
	// 			i
	// 				j (runnable)

	type root struct{}

	type a struct {
		Parent root `cli:"a,subcmd"`
	}

	type b struct {
		Parent root `cli:"b,subcmd"`
	}

	type c struct {
		Parent b `cli:"c,subcmd"`
	}

	type e struct {
		Parent b `cli:"e,subcmd"`
	}

	type f struct {
		Parent e `cli:"f,subcmd"`
	}

	type g struct {
		Parent root `cli:"g,subcmd"`
	}

	type h struct {
		Parent g `cli:"h,subcmd"`
	}

	type i struct {
		Parent g `cli:"i,subcmd"`
	}

	type j struct {
		Parent i `cli:"j,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ a) error { return nil },
		func(_ context.Context, _ b) error { return nil },
		func(_ context.Context, _ c) error { return nil },
		func(_ context.Context, _ f) error { return nil },
		func(_ context.Context, _ h) error { return nil },
		func(_ context.Context, _ j) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Command: command.Command{
			Config: reflect.TypeOf(root{}),
			Flags:  []command.Flag{helpFlag},
		},
		Children: map[string]cmdtree.ChildCommand{
			"a": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(a{}),
						Flags:  []command.Flag{helpFlag},
					},
				},
			},
			"b": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(b{}),
						Flags:  []command.Flag{helpFlag},
					},
					Children: map[string]cmdtree.ChildCommand{
						"c": cmdtree.ChildCommand{
							CommandTree: cmdtree.CommandTree{
								Command: command.Command{
									Config: reflect.TypeOf(c{}),
									Flags:  []command.Flag{helpFlag},
								},
							},
						},
						"e": cmdtree.ChildCommand{
							CommandTree: cmdtree.CommandTree{
								Command: command.Command{
									Config: reflect.TypeOf(e{}),
									Flags:  []command.Flag{helpFlag},
								},
								Children: map[string]cmdtree.ChildCommand{
									"f": cmdtree.ChildCommand{
										CommandTree: cmdtree.CommandTree{
											Command: command.Command{
												Config: reflect.TypeOf(f{}),
												Flags:  []command.Flag{helpFlag},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"g": cmdtree.ChildCommand{
				CommandTree: cmdtree.CommandTree{
					Command: command.Command{
						Config: reflect.TypeOf(g{}),
						Flags:  []command.Flag{helpFlag},
					},
					Children: map[string]cmdtree.ChildCommand{
						"h": cmdtree.ChildCommand{
							CommandTree: cmdtree.CommandTree{
								Command: command.Command{
									Config: reflect.TypeOf(h{}),
									Flags:  []command.Flag{helpFlag},
								},
							},
						},
						"i": cmdtree.ChildCommand{
							CommandTree: cmdtree.CommandTree{
								Command: command.Command{
									Config: reflect.TypeOf(i{}),
									Flags:  []command.Flag{helpFlag},
								},
								Children: map[string]cmdtree.ChildCommand{
									"j": cmdtree.ChildCommand{
										CommandTree: cmdtree.CommandTree{
											Command: command.Command{
												Config: reflect.TypeOf(j{}),
												Flags:  []command.Flag{helpFlag},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, removeFunc(tree))
}

func TestNew_MultipleRoots(t *testing.T) {
	type root1 struct{}
	type root2 struct{}

	_, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ root1) error { return nil },
		func(_ context.Context, _ root2) error { return nil },
	})

	assert.Equal(t,
		"multiple top-level commands: [cmdtree_test.root1 cmdtree_test.root2]",
		err.Error())
}

func removeFunc(tree cmdtree.CommandTree) cmdtree.CommandTree {
	tree.Command.Func = reflect.Value{}
	for k, v := range tree.Children {
		v.CommandTree = removeFunc(v.CommandTree)
		tree.Children[k] = v
	}

	return tree
}
