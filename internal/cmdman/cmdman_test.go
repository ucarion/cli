package cmdman_test

import (
	"context"
	"testing"

	"github.com/ucarion/cli/internal/cmdman"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdtree"
)

func TestMan_Basic(t *testing.T) {
	type args struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error { return nil },
	})

	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		"cmd.1": `.TH CMD 1
.SH NAME
cmd
.SH SYNOPSIS
\fIcmd\fR [<options>]
.SH DESCRIPTION

.SH OPTIONS
.TP
-h, --help
Display help message and exit.
`,
	}, cmdman.Man(tree, "./foo/bar/cmd"))
}

type gamutArgs struct {
	Foo string   `cli:"foo"`
	Bar string   `cli:"bar"`
	Baz []string `cli:"...baz"`
	X   string   `cli:"-x,--x-ray" value:"xxx" usage:"do some x stuff"`
	Y   string   `cli:"-y" value:"yyy" usage:"do some y stuff"`
	Z   *string  `cli:"--zulu" value:"zzz" usage:"do some z stuff"`
}

func (_ gamutArgs) Description() string {
	return "this is a short description"
}

func (_ gamutArgs) ExtendedDescription() string {
	return "this is an extended description"
}

func (_ gamutArgs) ExtendedUsage_X() string {
	return "this is extended x information"
}

func TestMan_Gamut(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ gamutArgs) error { return nil },
	})

	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		"cmd.1": `.TH CMD 1
.SH NAME
cmd - this is a short description
.SH SYNOPSIS
\fIcmd\fR [<options>] foo bar baz...
.SH DESCRIPTION
this is an extended description
.SH OPTIONS
.TP
-x, --x-ray <xxx>
this is extended x information
.TP
-y <yyy>

.TP
--zulu[=<zzz>]

.TP
-h, --help
Display help message and exit.
`,
	}, cmdman.Man(tree, "./foo/bar/cmd"))
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

func (_ rootArgs) Description() string {
	return "this is a short description"
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

	assert.Equal(t, map[string]string{
		"cmd-sub1.1": `.TH CMD-SUB1 1
.SH NAME
cmd-sub1
.SH SYNOPSIS
\fIcmd sub1\fR [<options>]
.SH DESCRIPTION

.SH OPTIONS
.TP
-h, --help
Display help message and exit.
`,
		"cmd-sub2.1": `.TH CMD-SUB2 1
.SH NAME
cmd-sub2
.SH SYNOPSIS
\fIcmd sub2\fR [<options>]
.SH DESCRIPTION

.SH OPTIONS
.TP
-h, --help
Display help message and exit.
`,
		"cmd.1": `.TH CMD 1
.SH NAME
cmd - this is a short description
.SH SYNOPSIS
\fIcmd\fR [<options>] [sub1 | sub2]
.SH DESCRIPTION
this is an extended description
.SH OPTIONS
.TP
-x, --x-ray <xxx>

.TP
-y <yyy>

.TP
--zulu[=<zzz>]

.TP
-h, --help
Display help message and exit.
`,
	}, cmdman.Man(tree, "./foo/bar/cmd"))
}

func TestHelp_NonExecutableWithSubcommands(t *testing.T) {
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ sub1Args) error { return nil },
		func(_ context.Context, _ sub2Args) error { return nil },
	})

	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		"cmd-sub1.1": `.TH CMD-SUB1 1
.SH NAME
cmd-sub1
.SH SYNOPSIS
\fIcmd sub1\fR [<options>]
.SH DESCRIPTION

.SH OPTIONS
.TP
-h, --help
Display help message and exit.
`,
		"cmd-sub2.1": `.TH CMD-SUB2 1
.SH NAME
cmd-sub2
.SH SYNOPSIS
\fIcmd sub2\fR [<options>]
.SH DESCRIPTION

.SH OPTIONS
.TP
-h, --help
Display help message and exit.
`,
		"cmd.1": `.TH CMD 1
.SH NAME
cmd - this is a short description
.SH SYNOPSIS
\fIcmd\fR [<options>] sub1 | sub2
.SH DESCRIPTION
this is an extended description
.SH OPTIONS
.TP
-x, --x-ray <xxx>

.TP
-y <yyy>

.TP
--zulu[=<zzz>]

.TP
-h, --help
Display help message and exit.
`,
	}, cmdman.Man(tree, "./foo/bar/cmd"))
}
