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

func main() {
	cli.Run(context.Background(), get, set)
}

type getArgs struct {
	RootArgs rootArgs `cli:"get,subcmd"`
	Key      string   `cli:"key"`
}

func get(ctx context.Context, args getArgs) error {
	fmt.Println("get", args)
	return nil
}

type setArgs struct {
	RootArgs rootArgs `cli:"set,subcmd"`
	Key      string   `cli:"key"`
	Value    string   `cli:"value"`
}

func set(ctx context.Context, args setArgs) error {
	fmt.Println("set", args)
	return nil
}
