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
	fmt.Println("get", args)
	return nil
}

func set(ctx context.Context, args setArgs) error {
	fmt.Println("set", args)
	return nil
}
