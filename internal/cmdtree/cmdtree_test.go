package cmdtree_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdtree"
)

func TestNew(t *testing.T) {
	// Verify that the simplest possible cmdtree creation works, and that the
	// given function is actually embedded properly in the returned value.
	t.Run("basic", func(t *testing.T) {
		type args struct {
		}

		called := false
		callErr := errors.New("dummy error")
		tree, err := cmdtree.New([]interface{}{
			func(ctx context.Context, a args) error {
				called = true
				assert.Equal(t, context.TODO(), ctx)
				assert.Equal(t, args{}, a)
				return callErr
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Config:   reflect.TypeOf(args{}),
			Children: []cmdtree.ChildCommand{},
		}, stripFuncs(tree))

		res := tree.Func.Call([]reflect.Value{
			reflect.ValueOf(context.TODO()),
			reflect.ValueOf(args{}),
		})

		assert.True(t, called)
		assert.Equal(t, res[0].Interface(), callErr)
	})

	t.Run("named root", func(t *testing.T) {
		type root struct {
			_ struct{} `cli:"root"`
		}

		tree, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ root) error {
				return nil
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Name:     "root",
			Config:   reflect.TypeOf(root{}),
			Children: []cmdtree.ChildCommand{},
		}, stripFuncs(tree))
	})

	t.Run("two subcommands", func(t *testing.T) {
		type root struct{}

		type foo struct {
			Root root `subcmd:"foo"`
			XXX  string
			YYY  string
		}

		type bar struct {
			XXX  string
			YYY  string
			Root root `subcmd:"bar"`
		}

		tree, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ foo) error {
				return nil
			},
			func(_ context.Context, _ bar) error {
				return nil
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Config: reflect.TypeOf(root{}),
			Children: []cmdtree.ChildCommand{
				cmdtree.ChildCommand{
					ParentConfigField: 0,
					CommandTree: cmdtree.CommandTree{
						Name:     "foo",
						Config:   reflect.TypeOf(foo{}),
						Children: []cmdtree.ChildCommand{},
					},
				},
				cmdtree.ChildCommand{
					ParentConfigField: 2,
					CommandTree: cmdtree.CommandTree{
						Name:     "bar",
						Config:   reflect.TypeOf(bar{}),
						Children: []cmdtree.ChildCommand{},
					},
				},
			},
		}, stripFuncs(tree))
	})

	t.Run("sub sub command", func(t *testing.T) {
		type root struct{}

		type foo struct {
			Root root `subcmd:"foo"`
			XXX  string
			YYY  string
		}

		type bar struct {
			XXX string
			YYY string
			Foo foo `subcmd:"bar"`
		}

		tree, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ bar) error {
				return nil
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Config: reflect.TypeOf(root{}),
			Children: []cmdtree.ChildCommand{
				cmdtree.ChildCommand{
					ParentConfigField: 0,
					CommandTree: cmdtree.CommandTree{
						Name:   "foo",
						Config: reflect.TypeOf(foo{}),
						Children: []cmdtree.ChildCommand{
							cmdtree.ChildCommand{
								ParentConfigField: 2,
								CommandTree: cmdtree.CommandTree{
									Name:     "bar",
									Config:   reflect.TypeOf(bar{}),
									Children: []cmdtree.ChildCommand{},
								},
							},
						},
					},
				},
			},
		}, stripFuncs(tree))
	})

	t.Run("sub sub command with runnable parent", func(t *testing.T) {
		type root struct{}

		type foo struct {
			Root root `subcmd:"foo"`
			XXX  string
			YYY  string
		}

		type bar struct {
			XXX string
			YYY string
			Foo foo `subcmd:"bar"`
		}

		tree, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ foo) error {
				return nil
			},
			func(_ context.Context, _ bar) error {
				return nil
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Config: reflect.TypeOf(root{}),
			Children: []cmdtree.ChildCommand{
				cmdtree.ChildCommand{
					ParentConfigField: 0,
					CommandTree: cmdtree.CommandTree{
						Name:   "foo",
						Config: reflect.TypeOf(foo{}),
						Children: []cmdtree.ChildCommand{
							cmdtree.ChildCommand{
								ParentConfigField: 2,
								CommandTree: cmdtree.CommandTree{
									Name:     "bar",
									Config:   reflect.TypeOf(bar{}),
									Children: []cmdtree.ChildCommand{},
								},
							},
						},
					},
				},
			},
		}, stripFuncs(tree))
	})

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
	t.Run("mixed runnable and sub-sub commands", func(t *testing.T) {
		type root struct{}

		type a struct {
			Parent root `subcmd:"a"`
		}

		type b struct {
			Parent root `subcmd:"b"`
		}

		type c struct {
			Parent b `subcmd:"c"`
		}

		type e struct {
			Parent b `subcmd:"e"`
		}

		type f struct {
			Parent e `subcmd:"f"`
		}

		type g struct {
			Parent root `subcmd:"g"`
		}

		type h struct {
			Parent g `subcmd:"h"`
		}

		type i struct {
			Parent g `subcmd:"i"`
		}

		type j struct {
			Parent i `subcmd:"j"`
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
			// root
			Config: reflect.TypeOf(root{}),
			Children: []cmdtree.ChildCommand{
				cmdtree.ChildCommand{
					CommandTree: cmdtree.CommandTree{
						Name:     "a",
						Config:   reflect.TypeOf(a{}),
						Children: []cmdtree.ChildCommand{},
					},
				},
				cmdtree.ChildCommand{
					CommandTree: cmdtree.CommandTree{
						Name:   "b",
						Config: reflect.TypeOf(b{}),
						Children: []cmdtree.ChildCommand{
							cmdtree.ChildCommand{
								CommandTree: cmdtree.CommandTree{
									Name:     "c",
									Config:   reflect.TypeOf(c{}),
									Children: []cmdtree.ChildCommand{},
								},
							},
							cmdtree.ChildCommand{
								CommandTree: cmdtree.CommandTree{
									Name:   "e",
									Config: reflect.TypeOf(e{}),
									Children: []cmdtree.ChildCommand{
										cmdtree.ChildCommand{
											CommandTree: cmdtree.CommandTree{
												Name:     "f",
												Config:   reflect.TypeOf(f{}),
												Children: []cmdtree.ChildCommand{},
											},
										},
									},
								},
							},
						},
					},
				},
				cmdtree.ChildCommand{
					CommandTree: cmdtree.CommandTree{
						Name:   "g",
						Config: reflect.TypeOf(g{}),
						Children: []cmdtree.ChildCommand{
							cmdtree.ChildCommand{
								CommandTree: cmdtree.CommandTree{
									Name:     "h",
									Config:   reflect.TypeOf(h{}),
									Children: []cmdtree.ChildCommand{},
								},
							},
							cmdtree.ChildCommand{
								CommandTree: cmdtree.CommandTree{
									Name:   "i",
									Config: reflect.TypeOf(i{}),
									Children: []cmdtree.ChildCommand{
										cmdtree.ChildCommand{
											CommandTree: cmdtree.CommandTree{
												Name:     "j",
												Config:   reflect.TypeOf(j{}),
												Children: []cmdtree.ChildCommand{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, stripFuncs(tree))
	})
}

func stripFuncs(tree cmdtree.CommandTree) cmdtree.CommandTree {
	children := []cmdtree.ChildCommand{}
	for _, c := range tree.Children {
		c.CommandTree = stripFuncs(c.CommandTree)
		children = append(children, c)
	}

	tree.Func = reflect.Value{}
	tree.Children = children
	return tree
}
