# cli

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/ucarion/cli)](https://pkg.go.dev/mod/github.com/ucarion/cli)

`github.com/ucarion/cli` is a Golang package for writing delightful, Unix-style
command-line tools in a type-safe way. With `github.com/ucarion/cli`, you can
define:

* Commands and sub-commands (`git commit`, `git remote`, `git remote set-url`)
* Short-style flags (`-f`, `-o json`, `-ojson`, `-abc`)
* Long-style flags (`--force`, `--output json`, `--output=json`)
* "Positional" arguments (`mv <from> <to>`, `cat <files...>`)

You will automatically get:

* `-h` and `--help` usage messages
* Man page generation (e.g. an automatically-generated `man my-cool-tool`)
* Bash and Zsh tab autocompletion (e.g. `mytool --f<TAB>` expands into `mytool
  --force`)

Best of all, `github.com/ucarion/cli` gives you all of this while keeping a
dirt-simple interface. Here's an unabridged, working tool built with `cli`:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	FirstName string `cli:"--first-name"`
	LastName  string `cli:"--last-name"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println("hello", args.FirstName, args.LastName)
		return nil
	})
}
```

This is [`examples/basic` in this repo](./examples/basic), which you can run as:

```text
$ go run ./examples/basic/... --first-name=john --last-name doe
hello john doe
```

## Installation

To use `cli` in your project, run:

```bash
go get github.com/ucarion/cli
```

## Demo

As an end-to-end demonstration of how you can use `cli` to build a tool with
subcommands, flags, arguments, `--help` text, `man` pages, and Bash/Zsh
completions, all with automated releases with GitHub actions and an
easy-to-install `brew` formula for macOS users, check out:

https://github.com/ucarion/fakegit

You can use `fakegit` to see what the most complex `cli` applications look like,
and you use it as a starting point in your own applications.

## Usage

For detailed, specific documentation on exactly what you can pass to `cli.Run`,
see the godocs for `github.com/ucarion/cli`. This section will work more as a
cookbook, showing you working programs that you can work off of.

At a high level, you use `cli` by passing `cli.Run` a context and a set of
functions. `cli` requires that every function you pass to `cli.Run` looks like:

```go
func (context.Context, T) error
```

Where `T` has to be a struct. `cli` will use reflection to determine the options
and arguments your command or sub-command expects. The rest of this section will
show examples of how you can use all of `cli'`s features.

### Accepting Options ("Flags")

To accept options (also called "flags"), mark a field in your struct with a tag
called `cli`. You can give your option a "short" name (e.g. `-f`), a "long" name
(e.g. `--force`), or both.

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Force   bool   `cli:"-f,--force"`
	Output  string `cli:"-o,--output"`
	N       int    `cli:"-n"`
	RFC3339 bool   `cli:"--rfc3339"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/options` in this repo](./examples/options), which you can run
as:

```text
$ go run ./examples/options/...
{false  0 false}

$ go run ./examples/options/... --force --output json --rfc3339 -n 5
{true json 5 true}
```

`cli` supports the full set of "standard" Unix command-line conventions, so this
also works, like it would with most tools in modern Linux distributions:

```text
$ go run ./examples/options/... -fn5 --rfc3339 --output=json
{true json 5 true}
```

### Accepting "Positional" Arguments

To accept arguments that aren't options, like the `pattern` and trailing list of
`files` in `grep pattern files...`, then tag your fields with `cli`, but don't
include a leading `-` or `--`. If your tag's value starts with `...`, then all
"leftover" / "trailing" arguments will go into that field.

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Foo string   `cli:"foo"`
	Bar string   `cli:"bar"`
	Baz []string `cli:"baz..."`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/posargs` in this repo](./examples/posargs), which you can run
as:

```text
$ go run ./examples/posargs/... a b
{a b []}

$ go run ./examples/posargs/... a b c d e
{a b [c d e]}
```

### Mixing Options and Arguments

As a relatively straightforward extension of the previous two examples, you can
accept both options ("flags") and "positional" arguments at once:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Force   bool     `cli:"-f,--force"`
	Output  string   `cli:"-o,--output"`
	N       int      `cli:"-n"`
	RFC3339 bool     `cli:"--rfc3339"`
	Foo     string   `cli:"foo"`
	Bar     string   `cli:"bar"`
	Baz     []string `cli:"baz..."`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/argsandopts` in this repo](./examples/argsandopts), which you
can run as:

```text
$ go run ./examples/argsandopts/... --force --output=json a b c d e --rfc3339
{true json 0 true a b [c d e]}
```

As is standard with Unix tools, `cli` will treat `--` in the input args as a
"end of flags" indicator. So, for instance, if you wanted `--rfc3339` above to
be treated as an argument instead of a flag, you could do:

```text
$ go run ./examples/argsandopts/... --force --output=json a b c d e -- --rfc3339
{true json 0 false a b [c d e --rfc3339]}
```

### Defining Commands and Sub-Commands

If you mark one of the fields of your struct like this:

```go
type bazArgs struct {
    ParentArgs parentArgs `cli:"baz,subcmd"`
}
```

That means that you're defining a sub-command called `baz`, and it's a
sub-command of the `parentArgs` type. When you pass a set of functions to
`cli.Run`, `cli` will use these `cli:"xxx,subcmd"` tags to discover your "tree"
of commands.

So, for instance, if you want to have a CLI tool that has a `get` and `set`
subcommands, you can do that like this:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type rootArgs struct {
	Username string `cli:"--username"`
	Password string `cli:"--password"`
}

type getArgs struct {
	RootArgs rootArgs `cli:"get,subcmd"`
	Key      string   `cli:"key"`
}

type setArgs struct {
	RootArgs rootArgs `cli:"set,subcmd"`
	Key      string   `cli:"key"`
	Value    string   `cli:"value"`
}

func main() {
	cli.Run(context.Background(), get, set)
}

func get(ctx context.Context, args getArgs) error {
	fmt.Println(args)
	return nil
}

func set(ctx context.Context, args setArgs) error {
	fmt.Println(args)
	return nil
}
```

This is [`examples/subcmds` in this repo](./examples/subcmds), which you can run
as:

```text
$ go run ./examples/subcmds/... --username foo --password bar get xxx
get {{foo bar} xxx}
$ go run ./examples/subcmds/... --username foo --password bar set xxx yyy
set {{foo bar} xxx yyy}
```

The pattern above of pointing to your parent config type via `cli:"xxx,subcmd"`
tags can work recursively. For instance, if you wanted to add `config get` and
`config set` subcommands to the above example, you could do:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type rootArgs struct {
	Username string `cli:"--username"`
	Password string `cli:"--password"`
}

type getArgs struct {
	RootArgs rootArgs `cli:"get,subcmd"`
	Key      string   `cli:"key"`
}

type setArgs struct {
	RootArgs rootArgs `cli:"set,subcmd"`
	Key      string   `cli:"key"`
	Value    string   `cli:"value"`
}

type configArgs struct {
	RootArgs   rootArgs `cli:"config,subcmd"`
	ConfigFile string   `cli:"--config-file"`
}

type getConfigArgs struct {
	ConfigArgs configArgs `cli:"get,subcmd"`
	Key        string     `cli:"key"`
}

type setConfigArgs struct {
	ConfigArgs configArgs `cli:"set,subcmd"`
	Key        string     `cli:"key"`
	Value      string     `cli:"value"`
}

func main() {
	cli.Run(context.Background(), get, set, getConfig, setConfig)
}

func get(ctx context.Context, args getArgs) error {
	fmt.Println("get", args)
	return nil
}

func set(ctx context.Context, args setArgs) error {
	fmt.Println("set", args)
	return nil
}

func getConfig(ctx context.Context, args getConfigArgs) error {
	fmt.Println("get config", args)
	return nil
}

func setConfig(ctx context.Context, args setConfigArgs) error {
	fmt.Println("set config", args)
	return nil
}
```

This is [`examples/nestedsubcmds` in this repo](./examples/nestedsubcmds), which
you can run as:

```text
$ go run ./examples/nestedsubcmds/... --username foo --password bar get xxx
get {{foo bar} xxx}
$ go run ./examples/nestedsubcmds/... --username foo --password bar set xxx yyy
set {{foo bar} xxx yyy}
$ go run ./examples/nestedsubcmds/... config --config-file=config.txt get xxx
get config {{{ } config.txt} xxx}
$ go run ./examples/nestedsubcmds/... config --config-file=config.txt set xxx yyy
set config {{{ } config.txt} xxx yyy}
```

You may notice that in the above example, `configArgs` is used as the parent
type to both `getConfigArgs` and `putConfigArgs`, but is never directly used by
any function you pass to `cli.Run`. When you do that, that indicates to `cli`
that you don't want the `config` subcommand to really "run". So this:

```text
$ go run ./examples/nestedsubcmds/... config
```

Just outputs help text, showing users that `config` takes a `--config-file`, and
that its subcommands are `get` and `set`:

```text
usage: /var/folders/.../exe/nestedsubcmds config [<options>] get|set

        --config-file <string>
    -h, --help                    display this help and exit
```

### Customizing Help Text

By default, `cli` will generate a help text for you, and it will be displayed if
the user passes `-h` or `--help`. By default, the help text looks like (see
[`examples/argsandopts` in this repo](./examples/argsandopts) for where these
flags and args are from):

```text
$ go run ./examples/argsandopts/... --help
usage: /var/folders/.../exe/argsandopts [<options>] foo bar baz...

    -f, --force
    -o, --output <string>
    -n <int>
        --rfc3339
    -h, --help               display this help and exit
```

The long `/var/folders/...` stuff is an artifact of how `go run` works, where it
first compiles the program into a temp directory with an esoteric name. `cli`'s
auto-generated help will figure out your program's name from `os.Args[0]`, as is
convention in Unix tools.

To get a less weird-looking name after `usage:` in the help text, try compiling
the program yourself first:

```text
$ go build ./examples/argsandopts/...
$ ./argsandopts --help
usage: ./argsandopts [<options>] foo bar baz...

    -f, --force
    -o, --output <string>
    -n <int>
        --rfc3339
    -h, --help               display this help and exit
```

There are a couple of things you can do to customize the help text:

* If you set a `ExtendedDescription() string` method on your args struct, then
  `cli` will call it, and use it as a description for your command.
* If you set a `usage` tag on a field, that will be shown next to the flag.
* If you set a `value` tag on a field, that will be shown instead of the
  `<string>` or `<int>` in the default output above.

Furthermore, if you define either `-h` or `--help` flag yourself, then `cli`
will leave it be. If you define both `-h` and `--help`, then `cli` will not show
help for you at all. This is useful mostly if you're writing a tool like `ls` or
`du`, where `-h` means "human-readable".

Putting all of that together, you can do:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Human   bool     `cli:"-h" usage:"show human-readable output"`
	Force   bool     `cli:"-f,--force" usage:"do the thing no matter what"`
	Output  string   `cli:"-o,--output" value:"format" usage:"the format to output in"`
	N       int      `cli:"-n" value:"times" usage:"how many times to do the thing"`
	RFC3339 bool     `cli:"--rfc3339" usage:"use rfc3339 timestamps"`
	Foo     string   `cli:"foo"`
	Bar     string   `cli:"bar"`
	Baz     []string `cli:"baz..."`
}

func (_ args) ExtendedDescription() string {
	return "This is just a program that shows you how to customize help text."
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/customhelptext` in this repo](./examples/customhelptext),
which you can run as:

```text
$ go run ./examples/customhelptext/... --help
usage: /var/folders/.../customhelptext [<options>] foo bar baz...

This is just a program that shows you how to customize help text.

    -h                       show human-readable output
    -f, --force              do the thing no matter what
    -o, --output <format>    the format to output in
    -n <times>               how many times to do the thing
        --rfc3339            use rfc3339 timestamps
        --help               display this help and exit
```

### Generating and Customizing Man Pages

If you set an environment variable called `UCARION_CLI_GENERATE_MAN`, then
`cli.Run` will generate `man` pages instead of running your program as usual.
The value of `UCARION_CLI_GENERATE_MAN` is the directory where the man pages
will be generated; each sub-command will get its own man page.

> Aside: it's called `UCARION_CLI_GENERATE_MAN` to make it more obvious what is
> reading the environment variable. The goal was to use a name that made it
> obvious that something called "ucarion cli" is generating a man page, which if
> you put in Google will hopefully lead you to the docs you are currently
> reading.

By default, the man pages look like (see [`examples/argsandopts` in this
repo](./examples/argsandopts) for where these flags and args are from):

```
$ UCARION_CLI_GENERATE_MAN="." go run ./examples/argsandopts/...
$ man ./argsandopts.1
```

```text
ARGSANDOPTS(1)                                                  ARGSANDOPTS(1)



NAME
       argsandopts

SYNOPSIS
       argsandopts [<options>] foo bar baz...

DESCRIPTION
OPTIONS
       -f, --force


       -o, --output <string>


       -n <int>


       --rfc3339


       -h, --help
              Display help message and exit.



                                                                ARGSANDOPTS(1)
```

There are a couple of things you can do to customize the help text:

* If you set a `Description() string` method on your args struct, then `cli`
  will call it, and the return value will appear after your program's name in
  the "Name" section.

  By convention, you should use a short, lower-case string for the description.
  For example, `ls`'s description is:

  ```text
  ls - list directory contents
  ```

* If you set a `ExtendedDescription() string` method on your args struct, then
  `cli` will call it, and use the return value as the "Description" for your
  command. This method is also used in help text, described in the previous
  section.

* If you set a `value` tag on a field, that will be shown instead of the
  `<string>` or `<int>` in the default output above.

* If you have a field called `XXX` in your struct (this is the "actual" name for
  the field not what you put in the `cli` tag), and if you have a method called
  `ExtendedUsage_XXX() string`, then `cli` will call it, and use the return
  value as the usage for the flag in man pages. This only applies to flags;
  there is no corresponding conventional way to describe "positional" arguments
  in man pages.

Putting all of that together, you can do:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Force   bool     `cli:"-f,--force"`
	Output  string   `cli:"-o,--output" value:"format"`
	N       int      `cli:"-n" value:"times"`
	RFC3339 bool     `cli:"--rfc3339"`
	Foo     string   `cli:"foo"`
	Bar     string   `cli:"bar"`
	Baz     []string `cli:"baz..."`
}

func (_ args) Description() string {
	return "dummy command with custom man page"
}

func (_ args) ExtendedDescription() string {
	return "This is just a program that shows you how to customize man pages."
}

func (_ args) ExtendedUsage_Force() string {
	return "Do the thing no matter what."
}

func (_ args) ExtendedUsage_Output() string {
	return "The format to output in."
}

func (_ args) ExtendedUsage_N() string {
	return "How many times to do the thing."
}

func (_ args) ExtendedUsage_RFC3339() string {
	return "Use RFC3339 timestamps."
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/custommanpage` in this repo](./examples/custommanpage), which
you can run as:

```text
$ UCARION_CLI_GENERATE_MAN="." go run ./examples/custommanpage/...
$ man ./custommanpage.1
```

```text
CUSTOMMANPAGE(1)                                              CUSTOMMANPAGE(1)



NAME
       custommanpage - dummy command with custom man page

SYNOPSIS
       custommanpage [<options>] foo bar baz...

DESCRIPTION
       This is just a program that shows you how to customize man pages.

OPTIONS
       -f, --force
              Do the thing no matter what.

       -o, --output <format>
              The format to output in.

       -n <times>
              How many times to do the thing.

       --rfc3339
              Use RFC3339 timestamps.

       -h, --help
              Display help message and exit.



                                                              CUSTOMMANPAGE(1)
```

### Generating Auto-Completions

The Bash and Zsh shells both support "completion" scripts that the shell will
run when you press "tab" (the Fish shell uses `man` pages to populate
completions, so the above section covers that) If you're not familiar with how
Bash/Zsh completion works, here's a crash course:

* You can register a completion script with Bash/Zsh using the builtin
  `complete`. In Bash, this builtin is available out of the box. In Zsh, you
  first have to run:

  ```sh
  autoload -U +X compinit && compinit
  autoload -U +X bashcompinit && bashcompinit
  ```

* Once you've registered a completion script, then when Bash/Zsh needs to
  generate completions, it will call the relevant completion script with the
  environment vars `COMP_LINE` (containing the line typed so far) and
  `COMP_CWORD` (containing the index of the word to complete).

Typically, programs that support completion ship with a Bash or Zsh script
alongside their main program, and they re-implement (a subset of) their flag
parsing in a shell script in order to generate completions.

`cli` takes a different approach. With `cli`, every program is its own
completion script. If `cli.Run` sees that the `COMP_LINE` and `COMP_CWORD`
environment variables are present, then `cli.Run` will output a set of
completions instead of running your program as usual.

For instance, to see what completions look like by default (see
[`examples/argsandopts` in this repo](./examples/argsandopts) for where these
flags and args are from):

```text
$ go build ./examples/argsandopts/...
$ complete -o bashdefault -o default -C ./argsandopts argsandopts
$ ./argsandopts -<TAB>
--force    --output   --rfc3339  -n
```

To emphasize how non-magic this is, you could also get those completions by
running `./argsandopts` yourself:

```text
$ COMP_LINE="./argsandopts -" COMP_CWORD="1" ./argsandopts
--force
--output
--rfc3339
-n
```

If your program has sub-commands, `cli.Run` will offer those sub-commands in its
completions. For instance (see [`examples/nestedsubcmds` in this
repo](./examples/nestedsubcmds) for where these commands and flags are from):

```text
$ ./nestedsubcmds <TAB>
--password  --username  config      get         set
$ ./nestedsubcmds config <TAB>
--config-file  get            set
```

### Customizing Auto-Completions

By default, `cli` will not offer any autocompletions for the value of a flag or
a positional argument. As a result of `-o` flags we passed to `complete` in the
previous section, Bash/Zsh will fall back to its default behavior, which is to
list files in the current directory:

```text
$ ./argsandopts --output <TAB>
README.md       argsandopts*    cli.go          ... (etc)
```

If you have a field called `XXX` in your struct (this is the "actual" name for
the field not what you put in the `cli` tag), and if you have a method called
`Autocomplete_XXX() []string`, then `cli` will call it, and will use the return
value as the suggested values for the flag or argument.

Crucially, your `Autocomplete_XXX` will be called *after* `cli` tries to parse
the flags the user has provided. That means that if your completions for a flag
or argument are a function of other flags, you can read those flag values to
figure out what to complete. This is especially useful if you have some sort of
`--config-file` or `--username`/`--password` flags that you need in order to
authenticate with a system, and then poll that system to figure out your
completions.

Putting all of that together, you can do:

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ucarion/cli"
)

type args struct {
	Foo string `cli:"--foo"`
	Bar string `cli:"--bar"`
}

func (a args) Autocomplete_Bar() []string {
	if a.Foo == "" {
		return nil
	}

	return []string{strings.ToUpper(a.Foo), strings.ToLower(a.Foo)}
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/customcompletions` in this
repo](./examples/customcompletions), which you can run as:

```text
$ go build ./examples/customcompletions/...
$ complete -o bashdefault -o default -C ./customcompletions customcompletions
$ ./customcompletions --foo hElLo --bar <TAB>
HELLO  hello
```

### Advanced Flag/Arg Use-Cases

This section will go through some more advanced use-cases for things you can do
with flags or arguments.

#### Passing a flag multiple times

If you mark a field as a flag, and that field's type is a slice (e.g.
`[]string`, `[]int`, etc.), then `cli` will let users pass that flag multiple
times. For example:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Names []string `cli:"--name"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/repeatedflag` in this repo](./examples/repeatedflag), which
you can run as:

```text
$ go run ./examples/repeatedflag/...
{[]}
$ go run ./examples/repeatedflag/... --name foo
{[foo]}
$ go run ./examples/repeatedflag/... --name foo --name bar --name baz
{[foo bar baz]}
```

#### Optionally-taking-value options

Some tools support options that can be provided either in the "boolean" way
(e.g. `mycmd --force`) or in the "takes-a-value" way (e.g. `mycmd
--output=json`). For instance, in `git` the `--color` flag, when it's supported,
can be provided with or without a value:

```bash
# These two do the same thing
git show HEAD --color
git show HEAD --color=auto

# This is different
git show HEAD --color=never
```

`cli` supports this use-case. If you mark a field as a flag, and that field's
type is a pointer (e.g. `*string`, `*int`, etc.), then `cli` will let users pass
that flag with or without a value.

If users don't pass the flag at all, the field will remain `nil` when it's
provided to you. If the users set the flag, but don't provide a value, then the
field will be instantiated as a pointer to the zero value of the type (e.g. for
`*string`, it would be a pointer to an empty string). If users set the flag and
provide a value, the field will be a pointer to that parsed value.

For example:

```go
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ucarion/cli"
)

type args struct {
	Color *string `cli:"--color"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		// We display this as json to avoid just printing a pointer here.
		return json.NewEncoder(os.Stdout).Encode(args)
	})
}
```

This is [`examples/optionallytakingvalue` in this
repo](./examples/optionallytakingvalue), which you can run as:

```text
$ go run ./examples/optionallytakingvalue/...
{"Color":null}
$ go run ./examples/optionallytakingvalue/... --color
{"Color":""}
$ go run ./examples/optionallytakingvalue/... --color=never
{"Color":"never"}
```

Optionally-taking-value options like this can be confusing to users. For
instance, this is not a valid invocation, because you're not allowed to put a
space between an optionally-taking-value option and its value:

```text
$ go run ./examples/optionallytakingvalue/... --color never
unexpected argument: never
```

What's going on here is that `cli`, in accordance with Unix convention, parses
`--color` as not having a value passed, and assumes `never` is a non-option
argument. But `examples/optionallytakingvalue` doesn't define any non-option
arguments, so `cli` reports an error to the user for the unexpected argument.

#### Custom parameter types

Out of the box, `cli` supports all of Go's number types (including floats, ints,
and units, but not complex numbers), as well as strings, for any option or
argument. If you'd like to parse options into a different type, you can:

1. Just do that parsing yourself, from within the function you pass to
   `cli.Run`, or
2. You define your own data type satisfying the `Param` interface in
   `github.com/ucarion/cli/param`. To satisfy that interface, you just need to
   support a `Set(string) error` method, where the `string` will be the user's
   input.

For example, here's how you can define a custom type of param that automatically
parses a URL:

```go
package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ucarion/cli"
)

type args struct {
	Foo urlParam `cli:"--foo"`
}

type urlParam struct {
	Value url.URL
}

func (p *urlParam) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	p.Value = *u
	return nil
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
```

This is [`examples/urlparam` in this repo](./examples/urlparam), which you can
run as:

```text
$ go run ./examples/urlparam/... --foo :
--foo: parse ":": missing protocol scheme
exit status 1
$ go run ./examples/urlparam/... --foo http://example.com/foo/bar
{{{http   example.com /foo/bar  false  }}}
```
