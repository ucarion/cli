package autocompleter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/autocompleter"
	"github.com/ucarion/cli/internal/cmdtree"
)

func TestAutocomplete_Basic(t *testing.T) {
	type args struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, []string(nil), autocompleter.Autocomplete(tree, nil))
}

func TestAutocomplete_Flags(t *testing.T) {
	type args struct {
		A string `cli:"-a"`
		B string `cli:"--bravo"`
		C int    `cli:"--charlie,-c"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"--bravo", "--charlie", "-a"},
		autocompleter.Autocomplete(tree, nil))

	// This is an invalid invocation. You can't set C (an int) to "xxx".
	assert.Equal(t, []string(nil), autocompleter.Autocomplete(tree, []string{"cmd", "-cxxx"}))
}

type autocompleteArgs struct {
	A string   `cli:"-a"`
	B string   `cli:"-b"`
	C string   `cli:"c"`
	D string   `cli:"d"`
	E []string `cli:"...e"`
}

func (a autocompleteArgs) Autocomplete_B() []string {
	return []string{a.A}
}

func (a autocompleteArgs) Autocomplete_C() []string {
	return []string{a.A}
}

func (a autocompleteArgs) Autocomplete_E() []string {
	return []string{"xxx", "yyy"}
}

func TestAutocomplete_FlagAutocomplete(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ autocompleteArgs) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"xxx"},
		autocompleter.Autocomplete(tree, []string{"cmd", "-axxx", "-b"}))
	assert.Equal(t,
		[]string(nil),
		autocompleter.Autocomplete(tree, []string{"cmd", "-a"}))
}

func TestAutocomplete_PosArgAutocomplete(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ autocompleteArgs) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"-b", "xxx"},
		autocompleter.Autocomplete(tree, []string{"cmd", "-axxx"}))
}

func TestAutocomplete_TrailingAutocomplete(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ autocompleteArgs) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"-a", "-b", "xxx", "yyy"},
		autocompleter.Autocomplete(tree, []string{"cmd", "a", "b"}))
}

type rootArgs struct {
	X string `cli:"-x"`
	Y string `cli:"-y"`
	Z string `cli:"-z"`
}

type sub1Args struct {
	Root rootArgs `cli:"sub1,subcmd"`
	A    string   `cli:"-a"`
}

func (_ sub1Args) Autocomplete_A() []string {
	return []string{"xxx"}
}

type sub2Args struct {
	Root rootArgs `cli:"sub2,subcmd"`
	B    string   `cli:"-b"`
}

func (_ sub2Args) Autocomplete_B() []string {
	return []string{"yyy"}
}

func TestAutocomplete_ExecutableWithSubcommands(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ sub1Args) error { return nil },
		func(_ context.Context, _ sub2Args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"-x", "-y", "-z", "sub1", "sub2"},
		autocompleter.Autocomplete(tree, []string{"cmd"}))
	assert.Equal(t,
		[]string{"-a"},
		autocompleter.Autocomplete(tree, []string{"cmd", "sub1"}))
	assert.Equal(t,
		[]string{"-b"},
		autocompleter.Autocomplete(tree, []string{"cmd", "sub2"}))
}
