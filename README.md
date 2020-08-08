# cli

`cli` is a package that lets you write delightful Unix-style CLI tools in a
type-safe way. In particular, `cli` supports:

* Subcommands (e.g. `git remote` or `fooctl get widget`)
* Positional arguments (e.g. `origin` and `master` in `git push origin master`)
* Short flag names (e.g. `-a -b -c 3` or `-abc 3`)
* Long flag names (e.g. `--dry-run` or `--foo=bar`)
* Generating `man` pages

You can think of `cli` as a type-safe, easier-to-use version of `cobra`. Here's
a working tool built with `cli`:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/ucarion/cli"
)

type args struct {
    Name string        `cli:"--name" help:"Who to greet"`
    Wait time.Duration `cli:"--duration" help:"How long to sleep before greeting"`
}

func main() {
    cli.Run(context.Background(), func(ctx context.Context, args args) error {
        time.Sleep(args.Wait)
        fmt.Println("hello", args.Name)

        return nil
    })
}
```
