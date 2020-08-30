package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

func main() {
	cli.Run(context.Background(), root, remoteUpdate)
}

type rootArgs struct {
	Version  bool   `cli:"--version"`
	Paginate bool   `cli:"-p,--paginate"`
	WorkTree string `cli:"--work-tree" value:"path"`
}

func (_ rootArgs) ExtendedDescription() string {
	return `Git is a fast, scalable, distributed revision control system with an unusually
rich command set that provides both high-level operations and full access to
internals.`
}

func root(ctx context.Context, args rootArgs) error {
	fmt.Printf("git: %+v\n", args)
	return nil
}

type remoteArgs struct {
	RootArgs rootArgs `cli:"remote,subcmd"`
	Verbose  bool     `cli:"-v,--verbose"`
}

type remoteUpdateArgs struct {
	RemoteArgs remoteArgs `cli:"update,subcmd"`
	Remote     string     `cli:"remote"`
}

func remoteUpdate(ctx context.Context, args remoteUpdateArgs) error {
	fmt.Printf("git remote update: %+v\n", args)
	return nil
}
