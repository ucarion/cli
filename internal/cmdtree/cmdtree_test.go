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
			Config: reflect.TypeOf(args{}),
		}, stripFuncs(tree))

		res := tree.Func.Call([]reflect.Value{
			reflect.ValueOf(context.TODO()),
			reflect.ValueOf(args{}),
		})

		assert.True(t, called)
		assert.Equal(t, res[0].Interface(), callErr)
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
			Root root `subcmd:"foo"`
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
					Child: cmdtree.CommandTree{
						Config:   reflect.TypeOf(foo{}),
						Children: []cmdtree.ChildCommand{},
					},
				},
				cmdtree.ChildCommand{
					ParentConfigField: 2,
					Child: cmdtree.CommandTree{
						Config:   reflect.TypeOf(bar{}),
						Children: []cmdtree.ChildCommand{},
					},
				},
			},
		}, stripFuncs(tree))
	})
}

func stripFuncs(tree cmdtree.CommandTree) cmdtree.CommandTree {
	children := []cmdtree.ChildCommand{}
	for _, c := range tree.Children {
		c.Child = stripFuncs(c.Child)
		children = append(children, c)
	}

	tree.Func = reflect.Value{}
	tree.Children = children
	return tree
}
