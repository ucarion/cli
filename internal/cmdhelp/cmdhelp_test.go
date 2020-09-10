package cmdhelp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdhelp"
	"github.com/ucarion/cli/internal/cmdtree"
)

func TestHelp_Basic(t *testing.T) {
	type args struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t, `usage: ./cmd [<options>]

    -h, --help    display this help and exit

`, cmdhelp.Help(tree, []string{"./cmd"}))
}

type gamutArgs struct {
	Foo string   `cli:"foo"`
	Bar string   `cli:"bar"`
	Baz []string `cli:"baz..."`
	X   string   `cli:"-x,--x-ray" value:"xxx" usage:"do some x stuff"`
	Y   string   `cli:"-y" value:"yyy" usage:"do some y stuff"`
	Z   *string  `cli:"--zulu" value:"zzz" usage:"do some z stuff"`
}

func (a gamutArgs) ExtendedDescription() string {
	return "this is an extended description"
}

func TestHelp_Gamut(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ gamutArgs) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		`usage: ./cmd [<options>] foo bar baz...

this is an extended description

    -x, --x-ray <xxx>     do some x stuff
    -y <yyy>              do some y stuff
        --zulu[=<zzz>]    do some z stuff
    -h, --help            display this help and exit

`, cmdhelp.Help(tree, []string{"./cmd"}))
}

type rootArgs struct {
	X string  `cli:"-x,--x-ray" value:"xxx" usage:"do some x stuff"`
	Y string  `cli:"-y" value:"yyy" usage:"do some y stuff"`
	Z *string `cli:"--zulu" value:"zzz" usage:"do some z stuff"`
}

type sub1Args struct {
	Root rootArgs `cli:"sub1,subcmd"`
}

type sub2Args struct {
	Root rootArgs `cli:"sub2,subcmd"`
}

func (_ rootArgs) ExtendedDescription() string {
	return "this is an extended description"
}

func TestHelp_ExecutableWithSubcommands(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ rootArgs) error { return nil },
		func(_ context.Context, _ sub1Args) error { return nil },
		func(_ context.Context, _ sub2Args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		`usage: ./cmd [<options>] [sub1|sub2]

this is an extended description

    -x, --x-ray <xxx>     do some x stuff
    -y <yyy>              do some y stuff
        --zulu[=<zzz>]    do some z stuff
    -h, --help            display this help and exit

`, cmdhelp.Help(tree, []string{"./cmd"}))
}

func TestHelp_NonExecutableWithSubcommands(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ sub1Args) error { return nil },
		func(_ context.Context, _ sub2Args) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		`usage: ./cmd [<options>] sub1|sub2

this is an extended description

    -x, --x-ray <xxx>     do some x stuff
    -y <yyy>              do some y stuff
        --zulu[=<zzz>]    do some z stuff
    -h, --help            display this help and exit

`, cmdhelp.Help(tree, []string{"./cmd"}))
}
