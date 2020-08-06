package cmd_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmd"
)

func TestFromFunc(t *testing.T) {
	// TestFromConfig handles ensuring the cli tags are properly parsed and
	// whatnot. This test mostly ensures that the function passed to FromFunc is
	// returned in the cmd's Fn.
	ok := false
	c, _, err := cmd.FromFunc(func(ctx context.Context, args struct{}) error {
		ok = true
		return nil
	})

	assert.NoError(t, err)

	reflect.ValueOf(c.Fn).Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(struct{}{}),
	})

	assert.True(t, ok)
}
