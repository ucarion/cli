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
	cli.Run(context.Background(), get, set, getConfig, setConfig)
}

type getArgs struct {
	RootArgs rootArgs `cli:"get,subcmd"`
	Key      string   `cli:"key"`
}

func get(ctx context.Context, args getArgs) error {
	fmt.Printf("get %#v\n", args)
	return nil
}

type setArgs struct {
	RootArgs rootArgs `cli:"set,subcmd"`
	Key      string   `cli:"key"`
	Value    string   `cli:"value"`
}

func set(ctx context.Context, args setArgs) error {
	fmt.Printf("set %#v\n", args)
	return nil
}

type configArgs struct {
	RootArgs   rootArgs `cli:"config,subcmd"`
	ConfigFile string   `cli:"--config-file"`
}

type getConfigArgs struct {
	ConfigArgs configArgs `cli:"get,subcmd"`
	Key        string     `cli:"key"`
}

func getConfig(ctx context.Context, args getConfigArgs) error {
	fmt.Printf("get config %#v\n", args)
	return nil
}

type setConfigArgs struct {
	ConfigArgs configArgs `cli:"set,subcmd"`
	Key        string     `cli:"key"`
	Value      string     `cli:"value"`
}

func setConfig(ctx context.Context, args setConfigArgs) error {
	fmt.Printf("set config %#v\n", args)
	return nil
}
