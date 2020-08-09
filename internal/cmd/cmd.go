package cmd

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Command struct {
	Flags  []Flag
	Config reflect.Type
	Func   reflect.Value
}

type Flag struct {
	LongNames  []string
	ShortNames []string
	ShortHelp  string
	Index      []int
	TakesValue bool
}

var typeBool = reflect.TypeOf(true)

func FromFunc(fn interface{}) (Command, error) {
	v := reflect.ValueOf(fn)
	config := v.Type().In(1)

	flags := []Flag{}
	for i := 0; i < config.NumField(); i++ {
		field := config.Field(i)

		cli := field.Tag.Get("cli")
		if cli == "" {
			continue
		}

		flag := Flag{
			LongNames:  []string{},
			ShortNames: []string{},
			ShortHelp:  field.Tag.Get("usage"),
			Index:      []int{i}, // TODO nested fields
			TakesValue: field.Type != typeBool,
		}

		for _, part := range strings.Split(cli, ",") {
			if strings.HasPrefix(part, "--") {
				flag.LongNames = append(flag.LongNames, part[2:])
			} else if strings.HasPrefix(part, "-") {
				// TODO assert there's exactly one char after the dash
				flag.ShortNames = append(flag.ShortNames, part[1:])
			}
		}

		flags = append(flags, flag)
	}

	return Command{Flags: flags, Config: config, Func: v}, nil
}

func (c *Command) Exec(ctx context.Context, argv []string) error {
	config := reflect.New(c.Config).Elem()

	for len(argv) > 0 {
		arg := argv[0]
		argv = argv[1:]

		if strings.HasPrefix(arg, "--") {
			// We are dealing with an argument like "--foo".

			// TODO support non-boolean flags.
			// TODO support flags that aren't found.
			flag, _ := c.findLongName(arg[2:])
			if err := flag.Set(config, ""); err != nil {
				return err
			}
		} else if strings.HasPrefix(arg, "-") {
			// We are dealing with an argument like "-f".

			arg = arg[1:] // strip off the leading "-"

			for len(arg) > 0 {
				char := arg[0:1]
				arg = arg[1:]

				// TODO support non-boolean flags.
				// TODO support flags that aren't found.
				flag, _ := c.findShortName(char)
				if err := flag.Set(config, ""); err != nil {
					return err
				}
			}

		}

		fmt.Println(arg, config)
	}

	// The config is now constructed. Let's call the underlying function with
	// the constructed config, and return any errors from that function.
	res := c.Func.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		config,
	})

	resErr := res[0].Interface()
	if resErr == nil {
		return nil
	}

	return resErr.(error)
}

func (c *Command) findLongName(s string) (Flag, bool) {
	for _, flag := range c.Flags {
		for _, longName := range flag.LongNames {
			if longName == s {
				return flag, true
			}
		}
	}

	return Flag{}, false
}

func (c *Command) findShortName(s string) (Flag, bool) {
	for _, flag := range c.Flags {
		for _, shortName := range flag.ShortNames {
			if shortName == s {
				return flag, true
			}
		}
	}

	return Flag{}, false
}

// Set identifies the field that f points to, and then parses s and sets the
// field to the parsed value within s.
//
// If f does not take a value, then s is ignored.
func (f Flag) Set(config reflect.Value, s string) error {
	field := config.FieldByIndex(f.Index)

	if !f.TakesValue {
		field.SetBool(true)
		return nil
	}

	switch v := field.Addr().Interface().(type) {
	case *string:
		*v = s
	case *time.Duration:
		dur, err := time.ParseDuration(s)
		if err != nil {
			return err
		}

		*v = dur
	}

	return nil
}
