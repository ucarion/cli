package cmd

import (
	"context"
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
			// We are dealing with an argument like "--foo" or "--foo=bar".
			var longName string // the "foo" in "--foo=bar"
			var value string    // the "bar" in "--foo=bar". may be empty
			if strings.ContainsRune(arg, '=') {
				// We are dealing with something like "--foo=bar". Split that
				// into "foo" and "bar".
				parts := strings.SplitN(arg, "=", 2)
				longName = parts[0][2:] // indexing [2:] is to strip out "--"
				value = parts[1]        // may include further "=" chars. That's ok
			} else {
				// We are dealing with something like "--foo". There is no
				// inline value.
				longName = arg[2:]
			}

			// TODO support flags that aren't found.
			flag, _ := c.findLongName(longName)
			if flag.TakesValue {
				if value == "" {
					// The flag takes a value, but it was not provided inline.
					// The next arg in argv is the value.
					//
					// TODO handle reaching end-of-argv
					value = argv[0]
					argv = argv[1:]
				}

				if err := flag.Set(config, value); err != nil {
					return err
				}
			} else {
				// The flag does not take a value. We can just set its value
				// right away.
				if err := flag.Set(config, ""); err != nil {
					return err
				}
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

				if flag.TakesValue {
					// The flag takes a value. There are three cases we might be
					// interested in:
					//
					// If the arg is like "-fbar", then we want to extract out
					// "bar" as an inline value to "-f".
					//
					// If the arg is like "-f=bar", then we want to extract out
					// "bar" as an inline value to "-f".
					//
					// If the arg is like "-f", then there is no inline value to
					// "-f". The next arg is its value.
					//
					// TODO support flags taking a value, but the value isn't
					// inline.
					var value string
					if strings.HasPrefix(arg, "=") {
						value = arg[1:]
					} else {
						value = arg
					}

					if value == "" {
						// The flag takes a value, but it was not provided
						// inline. The next arg in argv is the value.
						//
						// TODO handle reaching end-of-argv
						value = argv[0]
						argv = argv[1:]
					}

					if err := flag.Set(config, value); err != nil {
						return err
					}

					// empty these out so that the loop over arg terminates
					arg = ""
				} else {
					// The flag doesn't take a value. Just set this single flag,
					// and continue working on the argument.
					if err := flag.Set(config, ""); err != nil {
						return err
					}
				}
			}
		}
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
