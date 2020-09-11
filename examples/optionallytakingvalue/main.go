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
