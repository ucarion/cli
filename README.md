# cli

`github.com/ucarion/cli` is a package that helps you write quickly write
delightful and type-safe CLI tools. `cli` supports:

* Flags, like `--name foo`, `--id=bar`, or `--dry-run`
    * Flags can have shorthands, like `-n foo`, `-i=bar`, or `-d`
* Positional arguments, like the `2` and `32` in `pow 2 32`
    * Positional arguments can be fixed, like `repeat [string] [number]`
    * Positional arguments can be variadic, like `ls [file...]`
* Subcommands, like `git status` or `kubectl get pod`

There are plenty of other packages that give you these features. What stands out
about `cli` is that it gives you this with **perfect type safety**. `cli` infers
what flags you want from the argument of the functions you give it, and then
will parse results before giving them to you.

## Installation

To install this package, run:

```bash
go get github.com/ucarion/cli
```

## Usage

At its core, you use `cli` by giving it functions that take a `context.Context`
and a struct as an argument. `cli` will introspect that struct to figure out
what flags and positional arguments your command wants.

### Basic Usage (Positional Arguments)

To pass in arguments as positional arguments, set the `cli` tag on your struct
member to the name you of your positional argument you want to show in the
`--help` text:

```go
func main() {
    cli.Run(add)
}

// The order of the positional arguments is inferred from the order you name
// them here.
type args struct {
    A int `cli:"a"`
    B int `cli:"b"`
}

func add(ctx context.Context, a args) {
    fmt.Println(a.A + a.B)
}
```

If you invoke that function with `-h` or `--help`, you'll get (assuming you
compiled it into an executable called `mytool`):

...

### Basic Usage (Flags)

If you want to pass in arguments as flags, set the `cli` tag to a value that
starts with `--` (for verbose flag names) or `-` (for short flag names), or both
if you separate them by commas:

```go
func main() {
    cli.Run(add)
}

type args struct {
    A     int           `cli:"a"`
    B     int           `cli:"b"`
    Sleep time.Duration `cli:"--sleep,-s"`
}

func add(ctx context.Context, a args) error {
    time.Sleep(a.Sleep)
    fmt.Println(a.A + a.B)
    return nil
}
```

### Basic Usage (`--help` messages)

By default, all `cli` will show about a custom positional argument or flag its
name and type. If you want to show more information, you can use the `usage` tag
to control what `cli` should display in the `--help` message:

```go
func main() {
    cli.Run(add)
}

type args struct {
    A     int           `cli:"a" usage:"the first number to add"`
    B     int           `cli:"b" usage:"the second number to add"`
    Sleep time.Duration `cli:"--sleep,-s" usage:"how long to sleep before adding"`
}

func add(ctx context.Context, a args) {
    time.Sleep(a.Sleep)
    fmt.Println(a.A + a.B)
}
```

### Basic Usage (Default Values)

