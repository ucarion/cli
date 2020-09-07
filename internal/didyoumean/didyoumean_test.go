package didyoumean_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/didyoumean"
)

func TestDidYouMean(t *testing.T) {
	type rootArgs struct{}

	type sub1Args struct {
		Root rootArgs `cli:"aaa,subcmd"`
	}

	type sub2Args struct {
		Root rootArgs `cli:"bbb,subcmd"`
	}

	type sub3Args struct {
		Root rootArgs `cli:"ccc,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ sub1Args) error { return nil },
		func(_ context.Context, _ sub2Args) error { return nil },
		func(_ context.Context, _ sub3Args) error { return nil },
	})

	// These tests more or less take for granted that we have a working
	// levenstein distance implementation. All we're really trying to make sure
	// of here is that the module basically works.
	assert.NoError(t, err)
	assert.Equal(t, "aaa", didyoumean.DidYouMean(tree, "aab"))
	assert.Equal(t, "bbb", didyoumean.DidYouMean(tree, "bbz"))
	assert.Equal(t, "ccc", didyoumean.DidYouMean(tree, "cc"))
}
