// Package main is a program that fakes the behavior of a subset of the commands
// in git, to demonstrate most of cli's more advanced features.
package main

import (
	"context"

	"github.com/ucarion/cli"
)

func main() {
	cli.Run(
		checkout,
		add,
		commit,
		push,
	)
}

type rootArgs struct {
	_      struct{} `cli:"git"`
	Config []string `cli:"--config,-c"`
}

type checkoutArgs struct {
	_         rootArgs `cli:"checkout"`
	Quiet     bool     `cli:"--quiet,-q"`
	Force     bool     `cli:"--force,-f"`
	NewBranch string   `cli:"-b"`
	Branch    string   `cli:"branch"`
	Paths     []string `cli:"-- paths..."`
}

func checkout(_ context.Context, _ checkoutArgs) error {
	return nil
}

type addArgs struct {
	_ rootArgs `cli:"add"`
}

func add(_ context.Context, _ addArgs) error {
	return nil
}

type commitArgs struct {
	_ rootArgs `cli:"commit"`
}

func commit(_ context.Context, _ commitArgs) error {
	return nil
}

type pushArgs struct {
	_ rootArgs `cli:"push"`
}

func push(_ context.Context, _ pushArgs) error {
	return nil
}
