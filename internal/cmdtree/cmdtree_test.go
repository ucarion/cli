package cmdtree_test

import (
	"context"
	"errors"
	"reflect"
	"strconv"
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

	t.Run("flags and positional args", func(t *testing.T) {
		type embed3 struct {
			J string `cli:"-j"`
			K string `cli:"k"`
		}

		type embed2 struct {
			I string `cli:"-i"`
			embed3
		}

		type embed1 struct {
			H string `cli:"-h"`
		}

		type root struct {
			_ struct{} `cli:"root"`
			A string   `cli:"a"`
			B string   `cli:"-b"`
			C string   `cli:"--charlie"`
			embed1
			D string   `cli:"-d,--delta,-D,--dee"`
			E string   `cli:"echo"`
			F string   `cli:"foxtrot"`
			G []string `cli:"...golf"`
			embed2
		}

		tree, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ root) error {
				return nil
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, cmdtree.CommandTree{
			Name:   "root",
			Config: reflect.TypeOf(root{}),
			Flags: []cmdtree.Flag{
				cmdtree.Flag{Field: []int{2}, ShortNames: []string{"b"}, LongNames: []string{}},
				cmdtree.Flag{Field: []int{3}, ShortNames: []string{}, LongNames: []string{"charlie"}},
				cmdtree.Flag{Field: []int{4, 0}, ShortNames: []string{"h"}, LongNames: []string{}},
				cmdtree.Flag{Field: []int{5}, ShortNames: []string{"d", "D"}, LongNames: []string{"delta", "dee"}},
				cmdtree.Flag{Field: []int{9, 0}, ShortNames: []string{"i"}, LongNames: []string{}},
				cmdtree.Flag{Field: []int{9, 1, 0}, ShortNames: []string{"j"}, LongNames: []string{}},
			},
			PosArgs: []cmdtree.PosArg{
				cmdtree.PosArg{Field: []int{1}, Name: "a"},
				cmdtree.PosArg{Field: []int{6}, Name: "echo"},
				cmdtree.PosArg{Field: []int{7}, Name: "foxtrot"},
				cmdtree.PosArg{Field: []int{9, 1, 1}, Name: "k"},
			},
			TrailingArgs: cmdtree.PosArg{Field: []int{8}, Name: "golf"},
			Children:     []cmdtree.ChildCommand{},
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

	t.Run("gamut of valid types", func(t *testing.T) {
		type args struct {
			Bool          bool          `cli:"--bool"`
			Int           int           `cli:"--int"`
			Int8          int8          `cli:"--int8"`
			Int16         int16         `cli:"--int16"`
			Int32         int32         `cli:"--int32"`
			Int64         int64         `cli:"--int64"`
			Uint          uint          `cli:"--uint"`
			Uint8         uint8         `cli:"--uint8"`
			Uint16        uint16        `cli:"--uint16"`
			Uint32        uint32        `cli:"--uint32"`
			Uint64        uint64        `cli:"--uint64"`
			Float32       float32       `cli:"--float32"`
			Float64       float64       `cli:"--float64"`
			String        string        `cli:"--string"`
			Value         customValue   `cli:"--value"`
			StringArr     []string      `cli:"--string-arr"`
			ValueArr      []customValue `cli:"--value-arr"`
			OptionalValue *customValue  `cli:"--opt-value"`
			OptionalStr   *string       `cli:"--opt-string"`
		}

		_, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ args) error { return nil },
		})

		assert.NoError(t, err)
	})

	t.Run("not a valid function", func(t *testing.T) {
		type args struct{}

		testCases := []interface{}{
			nil,
			"foo",
			func() {},
			func(_ args) {},
			func(_ string, _ args) {},
			func(_ context.Context, _ args) {},
			func(_ args) error { return nil },
			func(_ context.Context, _ string) error { return nil },
			func(_ context.Context, _ *args) error { return nil },
			func(_ context.Context, _ args) string { return "" },
		}

		for i, tt := range testCases {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				_, err := cmdtree.New([]interface{}{tt})
				assert.Contains(t, err.Error(), "command funcs must be func(context.Context, T) error, got: ")
			})
		}
	})

	t.Run("multiple subcmd uses", func(t *testing.T) {
		type rootArgs struct{}

		type subArgs struct {
			Root1 rootArgs `subcmd:"foo"`
			Root2 rootArgs `subcmd:"foo"`
		}

		_, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ subArgs) error { return nil },
		})

		assert.Equal(t, "multiple uses of subcmd tag in config struct", err.Error())
	})

	t.Run("multiple root commands", func(t *testing.T) {
		type root1Args struct{}
		type root2Args struct{}

		_, err := cmdtree.New([]interface{}{
			func(_ context.Context, _ root1Args) error { return nil },
			func(_ context.Context, _ root2Args) error { return nil },
		})

		assert.Equal(t, "multiple top-level commands", err.Error())
	})

	t.Run("bad config field type", func(t *testing.T) {
		t.Run("string pointer pointer", func(t *testing.T) {
			type args struct {
				X **string `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported pointer param type: unsupported param type: *string",
				err.Error())
		})

		t.Run("channel", func(t *testing.T) {
			type args struct {
				X chan bool `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: chan bool",
				err.Error())
		})

		t.Run("uintptr", func(t *testing.T) {
			type args struct {
				X uintptr `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: uintptr",
				err.Error())
		})

		t.Run("complex64", func(t *testing.T) {
			type args struct {
				X complex64 `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: complex64",
				err.Error())
		})

		t.Run("func", func(t *testing.T) {
			type args struct {
				X func() `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: func()",
				err.Error())
		})

		t.Run("map", func(t *testing.T) {
			type args struct {
				X map[bool]bool `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: map[bool]bool",
				err.Error())
		})

		t.Run("interface", func(t *testing.T) {
			type args struct {
				X interface{} `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: interface {}",
				err.Error())
		})

		t.Run("struct", func(t *testing.T) {
			type args struct {
				X struct{} `cli:"-x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported param type: struct {}",
				err.Error())
		})

		t.Run("non-slice trailing args", func(t *testing.T) {
			type args struct {
				X string `cli:"...x"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: trailing args must be a slice",
				err.Error())
		})

		t.Run("bad config field in anonymous field", func(t *testing.T) {
			type embed struct {
				X **string `cli:"-x"`
			}

			type args struct {
				embed
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported pointer param type: unsupported param type: *string",
				err.Error())
		})

		t.Run("bad config field in parent type", func(t *testing.T) {
			type rootArgs struct {
				X **string `cli:"-x"`
			}

			type args struct {
				Root rootArgs `subcmd:"foo"`
			}

			_, err := cmdtree.New([]interface{}{
				func(_ context.Context, _ args) error { return nil },
			})

			assert.Equal(t,
				"X: unsupported pointer param type: unsupported param type: *string",
				err.Error())
		})
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

type customValue struct{}

func (c customValue) Set(_ string) error {
	return nil
}
