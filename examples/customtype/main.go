package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ucarion/cli"
)

type args struct {
	Foo bytes `cli:"--foo"`
}

type bytes int

func (b *bytes) UnmarshalText(text []byte) error {
	s := string(text)

	var base string
	var factor int

	switch {
	case strings.HasSuffix(s, "KB"):
		base = s[:len(s)-2]
		factor = 1024
	case strings.HasSuffix(s, "MB"):
		base = s[:len(s)-2]
		factor = 1024 * 1024
	case strings.HasSuffix(s, "GB"):
		base = s[:len(s)-2]
		factor = 1024 * 1024 * 1024
	case strings.HasSuffix(s, "TB"):
		base = s[:len(s)-2]
		factor = 1024 * 1024 * 1024 * 1024
	case strings.HasSuffix(s, "B"):
		base = s[:len(s)-1]
		factor = 1
	default:
		return fmt.Errorf("missing units suffix (must be one of B, KB, MB, GB, TB): %s", s)
	}

	n, err := strconv.ParseInt(base, 0, 0)
	if err != nil {
		return err
	}

	*b = bytes(int(n) * factor)
	return nil
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Printf("%#v\n", args)
		return nil
	})
}
