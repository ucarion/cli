package cmdtree_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmd"
	"github.com/ucarion/cli/internal/cmdtree"
)

func TestBasic(t *testing.T) {
	type rootArgs struct{}

	tree, err := cmdtree.FromFuncs([]interface{}{
		func(ctx context.Context, args rootArgs) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Root: cmd.Command{
			ConfigType: reflect.ValueOf(rootArgs{}).Type(),
			Flags:      []cmd.Flag{},
		},
		Children: map[string]cmdtree.CommandTree{},
	}, wipeFns(tree))
}

func TestTwoSubcommandsFromRoot(t *testing.T) {
	type rootArgs struct{}

	type fooArgs struct {
		rootArgs `cli:"foo"`
	}

	type barArgs struct {
		rootArgs `cli:"bar"`
	}

	tree, err := cmdtree.FromFuncs([]interface{}{
		func(ctx context.Context, args fooArgs) error {
			return nil
		},
		func(ctx context.Context, args barArgs) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Root: cmd.Command{
			Flags:      []cmd.Flag{},
			ConfigType: reflect.ValueOf(rootArgs{}).Type(),
		},
		Children: map[string]cmdtree.CommandTree{
			"foo": cmdtree.CommandTree{
				Root: cmd.Command{
					Name:       "foo",
					ConfigType: reflect.ValueOf(fooArgs{}).Type(),
					Flags:      []cmd.Flag{},
				},
				Children: map[string]cmdtree.CommandTree{},
			},
			"bar": cmdtree.CommandTree{
				Root: cmd.Command{
					Name:       "bar",
					ConfigType: reflect.ValueOf(barArgs{}).Type(),
					Flags:      []cmd.Flag{},
				},
				Children: map[string]cmdtree.CommandTree{},
			},
		},
	}, wipeFns(tree))
}

func TestNestedSubcommands(t *testing.T) {
	type rootArgs struct{}

	type fooArgs struct {
		rootArgs `cli:"foo"`
	}

	type barArgs struct {
		rootArgs `cli:"bar"`
	}

	type bazArgs struct {
		barArgs `cli:"baz"`
	}

	type quuxArgs struct {
		barArgs `cli:"quux"`
	}

	tree, err := cmdtree.FromFuncs([]interface{}{
		func(ctx context.Context, args fooArgs) error {
			return nil
		},
		func(ctx context.Context, args bazArgs) error {
			return nil
		},
		func(ctx context.Context, args quuxArgs) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Root: cmd.Command{
			Flags:      []cmd.Flag{},
			ConfigType: reflect.ValueOf(rootArgs{}).Type(),
		},
		Children: map[string]cmdtree.CommandTree{
			"foo": cmdtree.CommandTree{
				Root: cmd.Command{
					Name:       "foo",
					ConfigType: reflect.ValueOf(fooArgs{}).Type(),
					Flags:      []cmd.Flag{},
				},
				Children: map[string]cmdtree.CommandTree{},
			},
			"bar": cmdtree.CommandTree{
				Root: cmd.Command{
					Name:       "bar",
					ConfigType: reflect.ValueOf(barArgs{}).Type(),
					Flags:      []cmd.Flag{},
				},
				Children: map[string]cmdtree.CommandTree{
					"baz": cmdtree.CommandTree{
						Root: cmd.Command{
							Name:       "baz",
							ConfigType: reflect.ValueOf(bazArgs{}).Type(),
							Flags:      []cmd.Flag{},
						},
						Children: map[string]cmdtree.CommandTree{},
					},
					"quux": cmdtree.CommandTree{
						Root: cmd.Command{
							Name:       "quux",
							ConfigType: reflect.ValueOf(quuxArgs{}).Type(),
							Flags:      []cmd.Flag{},
						},
						Children: map[string]cmdtree.CommandTree{},
					},
				},
			},
		},
	}, wipeFns(tree))
}

func TestDeeplyNestedSubcommands(t *testing.T) {
	type rootArgs struct{}

	type fooArgs struct {
		rootArgs `cli:"foo"`
	}

	type barArgs struct {
		fooArgs `cli:"bar"`
	}

	type bazArgs struct {
		barArgs `cli:"baz"`
	}

	type quuxArgs struct {
		bazArgs `cli:"quux"`
	}

	tree, err := cmdtree.FromFuncs([]interface{}{
		func(ctx context.Context, args quuxArgs) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, cmdtree.CommandTree{
		Root: cmd.Command{
			Flags:      []cmd.Flag{},
			ConfigType: reflect.ValueOf(rootArgs{}).Type(),
		},
		Children: map[string]cmdtree.CommandTree{
			"foo": cmdtree.CommandTree{
				Root: cmd.Command{
					Name:       "foo",
					ConfigType: reflect.ValueOf(fooArgs{}).Type(),
					Flags:      []cmd.Flag{},
				},
				Children: map[string]cmdtree.CommandTree{
					"bar": cmdtree.CommandTree{
						Root: cmd.Command{
							Name:       "bar",
							ConfigType: reflect.ValueOf(barArgs{}).Type(),
							Flags:      []cmd.Flag{},
						},
						Children: map[string]cmdtree.CommandTree{
							"baz": cmdtree.CommandTree{
								Root: cmd.Command{
									Name:       "baz",
									ConfigType: reflect.ValueOf(bazArgs{}).Type(),
									Flags:      []cmd.Flag{},
								},
								Children: map[string]cmdtree.CommandTree{
									"quux": cmdtree.CommandTree{
										Root: cmd.Command{
											Name:       "quux",
											ConfigType: reflect.ValueOf(quuxArgs{}).Type(),
											Flags:      []cmd.Flag{},
										},
										Children: map[string]cmdtree.CommandTree{},
									},
								},
							},
						},
					},
				},
			},
		},
	}, wipeFns(tree))
}

func TestMultipleRoots(t *testing.T) {
	type fooArgs struct{}
	type barArgs struct{}

	_, err := cmdtree.FromFuncs([]interface{}{
		func(ctx context.Context, args fooArgs) error {
			return nil
		},
		func(ctx context.Context, args barArgs) error {
			return nil
		},
	})

	assert.Equal(t, err, cmdtree.ErrMultipleRoots)
}

func wipeFns(tree cmdtree.CommandTree) cmdtree.CommandTree {
	tree.Root.Fn = nil
	for k, v := range tree.Children {
		tree.Children[k] = wipeFns(v)
	}

	return tree
}
