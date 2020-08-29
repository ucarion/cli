package tagparse

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Kind string

const (
	KindSubcmd = "subcmd"
	KindFlag   = "flag"
	KindPosArg = "posarg"
)

type ParsedTag struct {
	Kind          Kind
	CommandName   string
	ShortFlagName string
	LongFlagName  string
	FlagValueName string
	PosArgName    string
	IsTrailing    bool
	Usage         string
}

const (
	tagCLI   = "cli"
	tagValue = "value"
	tagUsage = "usage"

	cliSubcmd = "subcmd"
)

var paramRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")

func Parse(tag reflect.StructTag) (ParsedTag, error) {
	cli, ok := tag.Lookup(tagCLI)
	if !ok {
		return ParsedTag{}, nil
	}

	cliParts := strings.Split(cli, ",")
	if len(cliParts) > 2 {
		return ParsedTag{}, fmt.Errorf("too many options in cli tag: %v", cli)
	}

	var parsed ParsedTag
	if len(cliParts) > 1 && cliParts[1] == cliSubcmd {
		// We are dealing with a subcmd-kinded tag.
		if !paramRegex.MatchString(cliParts[0]) {
			return ParsedTag{}, fmt.Errorf("invalid subcommand name: %v", cliParts[0])
		}

		parsed = ParsedTag{Kind: KindSubcmd, CommandName: cliParts[0]}
	} else if strings.HasPrefix(cliParts[0], "-") {
		// We are dealing with a flag-kinded tag. The two parts must both be
		// flags, and cannot both be short or both be long. Only long flags can
		// begin with two dashes.
		var shortName, longName string

		for _, part := range cliParts {
			if strings.HasPrefix(part, "--") {
				if longName != "" {
					return ParsedTag{}, fmt.Errorf("flags can only have one long form: %v", part)
				}

				if !paramRegex.MatchString(part[2:]) {
					return ParsedTag{}, fmt.Errorf("invalid long flag name: %v", part)
				}

				longName = part[2:]
			} else if strings.HasPrefix(part, "-") {
				if shortName != "" {
					return ParsedTag{}, fmt.Errorf("flags can only have one short form: %v", part)
				}

				if !paramRegex.MatchString(part[1:]) || len(part) != 2 {
					return ParsedTag{}, fmt.Errorf("invalid short flag name: %v", part)
				}

				shortName = part[1:]
			} else {
				return ParsedTag{}, fmt.Errorf("invalid flag name: %v", part)
			}
		}

		parsed = ParsedTag{Kind: KindFlag, ShortFlagName: shortName, LongFlagName: longName}
	} else if strings.HasPrefix(cliParts[0], "...") {
		// We are dealing with a trailing posarg-kinded tag.
		if !paramRegex.MatchString(cliParts[0][3:]) {
			return ParsedTag{}, fmt.Errorf("invalid positional argument name: %v", cliParts[0])
		}

		parsed = ParsedTag{Kind: KindPosArg, PosArgName: cliParts[0][3:], IsTrailing: true}
	} else {
		// We are dealing with a posarg-kinded tag.
		if !paramRegex.MatchString(cliParts[0]) {
			return ParsedTag{}, fmt.Errorf("invalid positional argument name: %v", cliParts[0])
		}

		parsed = ParsedTag{Kind: KindPosArg, PosArgName: cliParts[0]}
	}

	parsed.FlagValueName = tag.Get(tagValue)
	parsed.Usage = tag.Get(tagUsage)
	return parsed, nil
}
