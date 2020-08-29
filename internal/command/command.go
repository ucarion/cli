package command

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ucarion/cli/internal/tagparse"
	"github.com/ucarion/cli/param"
)

type Command struct {
	Func                reflect.Value
	Config              reflect.Type
	Description         string
	ExtendedDescription string
	Flags               []Flag
	PosArgs             []PosArg
	Trailing            PosArg
}

type Flag struct {
	ShortName     string
	LongName      string
	Usage         string
	ExtendedUsage string
	ValueName     string
	IsHelp        bool
	FieldIndex    []int
}

type PosArg struct {
	Name       string
	FieldIndex []int
}

type ParentInfo struct {
	ChildName          string
	ParentType         reflect.Type
	ParentIndexInChild int
}

type description interface {
	Description() string
}

type extendedDescription interface {
	ExtendedDescription() string
}

const extendedUsagePrefix = "ExtendedUsage_"

func FromFunc(fn interface{}) (Command, ParentInfo, error) {
	t := reflect.TypeOf(fn)

	if err := checkValidFunc(t); err != nil {
		return Command{}, ParentInfo{}, err
	}

	cmd, pinfo, err := FromType(t.In(1))
	cmd.Func = reflect.ValueOf(fn)
	return cmd, pinfo, err
}

var (
	paramType = reflect.TypeOf((*param.Param)(nil)).Elem()
	ctxType   = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType   = reflect.TypeOf((*error)(nil)).Elem()
)

func checkValidFunc(t reflect.Type) error {
	err := fmt.Errorf("command funcs must have type: func(context.Context, T) error, got: %v", t)

	if t == nil {
		return err
	}

	if t.Kind() != reflect.Func {
		return err
	}

	if t.NumIn() != 2 {
		return err
	}

	if !t.In(0).Implements(ctxType) {
		return err
	}

	if t.In(1).Kind() != reflect.Struct {
		return err
	}

	if t.NumOut() != 1 {
		return err
	}

	if !t.Out(0).Implements(errType) {
		return err
	}

	return nil
}

func FromType(t reflect.Type) (Command, ParentInfo, error) {
	cmd := Command{Config: t}
	var pinfo ParentInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag, err := tagparse.Parse(f.Tag)
		if err != nil {
			return Command{}, ParentInfo{}, err
		}

		if tag.Kind == tagparse.KindSubcmd {
			pinfo = ParentInfo{
				ChildName:          tag.CommandName,
				ParentType:         f.Type,
				ParentIndexInChild: i,
			}
		}
	}

	v := reflect.New(t).Interface()

	if v, ok := v.(description); ok {
		cmd.Description = v.Description()
	}

	if v, ok := v.(extendedDescription); ok {
		cmd.ExtendedDescription = v.ExtendedDescription()
	}

	if err := addParams(&cmd, nil, t); err != nil {
		return Command{}, ParentInfo{}, err
	}

	addHelpFlag(&cmd)

	return cmd, pinfo, nil
}

func addParams(cmd *Command, index []int, t reflect.Type) error {
	v := reflect.Zero(t)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Anonymous {
			if err := addParams(cmd, append(index, i), f.Type); err != nil {
				return err
			}
		}

		tag, err := tagparse.Parse(f.Tag)
		if err != nil {
			return err
		}

		switch tag.Kind {
		case tagparse.KindFlag:
			// Ensure the field is a valid param.
			if _, err := param.New(reflect.New(f.Type).Interface()); err != nil {
				return fmt.Errorf("%v: %w", f.Name, err)
			}

			var extendedUsage string
			if m, ok := t.MethodByName(extendedUsagePrefix + f.Name); ok {
				// Ensure the method has the right signature: it takes in a
				// receiver and no args, and returns just a string.
				if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0) == reflect.TypeOf("") {
					m := v.MethodByName(extendedUsagePrefix + f.Name)
					extendedUsage = m.Call(nil)[0].Interface().(string)
				}
			}

			cmd.Flags = append(cmd.Flags, Flag{
				ShortName:     tag.ShortFlagName,
				LongName:      tag.LongFlagName,
				Usage:         tag.Usage,
				ExtendedUsage: extendedUsage,
				ValueName:     tag.FlagValueName,
				FieldIndex:    append(index, i),
			})
		case tagparse.KindPosArg:
			// Ensure the field is a valid param.
			if _, err := param.New(reflect.New(f.Type).Interface()); err != nil {
				return fmt.Errorf("%v: %w", f.Name, err)
			}

			posArg := PosArg{
				Name:       tag.PosArgName,
				FieldIndex: append(index, i),
			}

			if tag.IsTrailing {
				cmd.Trailing = posArg
			} else {
				cmd.PosArgs = append(cmd.PosArgs, posArg)
			}
		}
	}

	return nil
}

const (
	shortHelp         = "h"
	longHelp          = "help"
	helpUsage         = "display this help and exit"
	helpExtendedUsage = "Display help message and exit."
)

func addHelpFlag(cmd *Command) {
	helpFlag := Flag{
		IsHelp:        true,
		ShortName:     shortHelp,
		LongName:      longHelp,
		Usage:         helpUsage,
		ExtendedUsage: helpExtendedUsage,
	}

	for _, f := range cmd.Flags {
		if f.ShortName == shortHelp {
			helpFlag.ShortName = ""
		}

		if f.LongName == longHelp {
			helpFlag.LongName = ""
		}
	}

	if helpFlag.ShortName == "" && helpFlag.LongName == "" {
		return
	}

	cmd.Flags = append(cmd.Flags, helpFlag)
}
