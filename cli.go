// Package cli implements Unix-style arg parsing and evaluation.
//
// The documentation here describes the contract that cli upholds. For more
// high-level "cookbook"-style documentation, see the cli README, available
// online at:
//
// https://github.com/ucarion/cli
package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ucarion/cli/internal/autocompleter"
	"github.com/ucarion/cli/internal/cmdman"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/exectree"
)

const (
	envCompleteLine = "COMP_LINE"
	envCompleteArgc = "COMP_CWORD"

	envGenerateManDir = "UCARION_CLI_GENERATE_MAN"
)

// Run constructs and executes a command tree from a set of functions.
//
// Command Trees
//
// A command tree is cli's representation of a CLI application. The funcs passed
// to Run are used to construct a command tree; if the funcs cannot be
// constructed into a command tree, then Run panics.
//
// Each of the values in funcs must be of the type:
//
//  func(context.Context, T) error
//
// Where T is a struct, called a "config struct".
//
// Run constructs a directed graph of such config structs, where "child" config
// structs point to their single "parent" config struct, and also keep track of
// the child's name. Config structs indicate their parent and name using the cli
// tag, described further below.
//
// There must be exactly one config struct that does not have a parent config
// struct; this config struct is the "root" of the command-tree. The root config
// struct is not explicitly named; its name is derived from the first element of
// os.Args.
//
// Run discovers the set of config structs by first using each of the second
// arguments to the elements of funcs as an entry-point. Run will then
// iteratively walk the set of config structs, discovering a full command tree.
// In other words: for a config struct to be part of the command tree, it either
// needs to be directly used by one of the elements of funcs, or it needs to be
// used reachable by following the parents of one of those directly-used
// structs.
//
// If a config struct is directly used by one of the elements of funcs, then it
// is said to be "runnable". All other config structs are said to be
// "non-runnable".
//
// To illustrate these concepts with an example, consider a tool like a subset
// of git, the version control tool. These are all valid invocations of "git":
//
//  git commit
//  git remote
//  git remote add
//  git worktree add
//
// But these are not valid invocations (they will output a help message instead
// of actually performing any action):
//
//  git
//  git worktree
//
// In cli's model, you can represent this as a command-tree like so:
//
//  git (non-runnable):
//      commit (runnable)
//      remote (runnable):
//          add (runnable)
//      worktree (non-runnable):
//          add (runnable)
//
// Config Structs
//
// The previous section describes how config structs are discovered. This
// section will describe how to define a config struct, and elaborate on the
// data that can be associated with a config struct.
//
// A config struct must be a Go struct. Each config struct may define a set of
// options (the stdlib flag package calls these "flags"), arguments (sometimes
// called "positional arguments"), and a single set of "trailing arguments".
// Additionally, each config struct may have up to one "parent" config type; if
// a config struct has a parent config type, it must also have a name. All of
// these attributes are defined using the struct tag with the key "cli".
//
// If a field of a config struct is not tagged with "cli", then Run will ignore
// it. The "cli" tag may be used in a few forms.
//
// The parent form is indicated by setting "cli" to "xxx,subcmd", where "xxx" is
// the name of the sub-command. The type of the field is the parent config type.
// For instance, this indicates that a config struct is named "bar", and its
// parent type is parentConfigType:
//
//  Foo parentConfigStruct `cli:"bar,subcmd"`
//
// The option form is indicated by setting "cli" to one of "-x", "--yyy", or
// "-x,--yyy", where "x" is the "short" name of the option and "yyy" is the
// "long" name of the option. For example:
//
//  // This is an option with only a short name.
//  Force bool `cli:"-f"`
//
//  // This is an option with only a long name.
//  RFC3339 bool `cli:"--rfc3339"`
//
//  // This is an option with both a short name and a long name.
//  Verbose bool `cli:"-v,--verbose"
//
// The argument form is indicated by setting "cli" to "xxx", where "xxx" is the
// name of the argument. For example:
//
//  // This is a argument called "path".
//  Path string `cli:"path"`
//
// The trailing argument form is indicated by setting "cli" to "xxx...", where
// "xxx" is the name of the trailing arguments. For example:
//
//  // This is a set of trailing arguments called "files"
//  Files []string `cli:"files..."`
//
// Put together, options, arguments, and trailing arguments construct a data
// model familiar to users of Unix-like tools. For instance, if you have a tool
// which you can invoke as (where "[...]" means something is optional):
//
//  mytool [-f] [--rfc3339] [-v | --verbose] path [file1 [file2 [file3 ...]]]
//
// You could represent this with "cli" tags as:
//
//  type args struct {
//      Force   bool     `cli:"-f"`
//      RFC3339 bool     `cli:"--rfc3339"`
//      Verbose bool     `cli:"-v,--verbose"
//      Path    string   `cli:"path"`
//      Files   []string `cli:"files..."`
//  }
//
// If "mytool" were a subcommand of some larger "megatool", whose config type is
// megaArgs, then you could represent that relationship with an additional
// field:
//
//  type args struct {
//      ParentArgs megaArgs `cli:"mytool,subcmd"`
//      // ...
//  }
//
// Any field that uses the "cli" tag may also use the "usage" tag. That tag's
// value is set to be the "usage" attribute of the option. The "usage" tag has
// no use on fields that are not options.
//
// Any field that uses the "cli" tag may also use the "value" tag. That tag's
// value is set to be the "value name" of the option. The "value" tag has no use
// on fields that are not options that take value.
//
// If a "cli"-using field named "XXX" has a corresponding method named
// "ExtendedUsage_XXX" on the struct with the signature:
//
//  func() string
//
// Then Run will call that function, and the return value is set to be the
// "extended usage" attribute of the option.
//
// If a "cli"-using field named "XXX" has a corresponding method named
// "Autocomplete_XXX" on the struct with the signature:
//
//  func() []string
//
// Then that method is set to be the "autocompleter" attribute of the option.
//
// If a config type satisfies satisfies this interface:
//
//  interface {
//      Description() string
//  }
//
// Then Run will call Description, and the return value will be used as the
// "description" attribute of the command.
//
// If a config type satisfies satisfies this interface:
//
//  interface {
//      ExtendedDescription() string
//  }
//
// Then Run will call ExtendedDescription, and the return value will be used as
// the "extended description" attribute of the command.
//
// These "usage", "value name", "extended usage", "autocompleter",
// "description", and "extended description" attributes are later used in the
// "Command-Line Argument Parsing", "Man Page Generation", and "Bash/Zsh
// Completions" sections below.
//
// By default, all config types are implicitly populated by an additional option
// whose short name is "h" and whose long name is "help", unless those names are
// already specified. This additional option is internally marked as being a
// special help option; how this affects Run is covered further in "Command-Line
// Argument Parsing" below.
//
// Config types may embed other structs. Any options, arguments, or trailing
// arguments defined within those embedded structs will be honored. However: the
// parent form of the "cli" tag will be ignored if it's placed within a embedded
// struct.
//
// It's recommended that you use embedded structs to reduce code duplication if
// you have the same sets of options appear in many different commands in your
// tool.
//
// Parameter Types
//
// In the above examples, the fields were of type string and bool. This section
// will describe what other types are permitted as values for options,
// arguments, and trailing arguments.
//
// Run supports all of the following types as field types out of the box:
//
//  bool
//  byte
//  int
//  uint
//  rune
//  string
//  int8
//  uint8
//  int16
//  uint16
//  int32
//  uint32
//  int64
//  uint64
//  float32
//  float64
//
// For all of the types above (except string), values are parsed from os.Args
// using the appropriate method from the strconv package in the standard
// library.
//
// Run also works with any type that implements TextUnmarshaler from the
// encoding standard library package.
//
// Run will call TextUnmarshal on the zero value of your TextUnmarshal
// implementation, where the text is the value to parse. If TextUnmarshal
// returns an error, then the text is considered a bad argument.
//
// Furthermore, all of the types above are supported in slices or pointers. In
// other words, if T is one of the types described previously (it is a
// TextUnmarshaler or is in the list of primitive types above), then []T and *T
// are supported as well. This rule does not apply recursively; [][]T is not
// supported.
//
// Wrapping a type with a slice (that is, doing "[]T") indicates that the
// argument can be passed multiple times. For trailing arguments, the type must
// be wrapped in a slice; this is because trailing arguments must, by
// definition, support being passed multiple times.
//
// Wrapping a type with a pointer (that is, doing "*T") indicates that the
// argument may, but does not have to, take a value. See below for more details
// on how command-line argument parsing works.
//
// Command-Line Argument Parsing
//
// If the COMP_LINE, COMP_CWORD, and UCARION_CLI_GENERATE_MAN environment
// variables are not populated, then Run will parse os.Args against the command
// tree formed from funcs, construct an instance of the appropriate config type
// with fields populated from the parsed args, and then call the appropriate
// function in funcs with ctx and the constructed config type.
//
// At a high level, the syntax that Run expects from os.Args like this:
//
//  root-name [root-options] subcmd [subcmd-options] args...
//
// In other words: Run expects the root name of the command in os.Args[0]. It
// expects that options for a command or subcommand go immediately after the
// command or subcommand's name. In other words, flags do not "propagate" or get
// "inherited". For example, if the root-level command takes an option "-x", and
// a subcommand "y" takes some option "-z", then this is a valid invocation:
//
//  cmd -x y -z
//
// But this is not:
//
//  cmd y -x -z
//
// Because the options for the root-level command must go before the
// sub-command's name.
//
// Run follows the conventions established by GNU's extensions to the getopt
// standard from POSIX; these are the conventions familiar to users of most
// modern Linux distributions. In particular:
//
// Options are always optional. Run implements this convention by defaulting all
// options to their zero value.
//
// Non-trailing arguments are never optional. Trailing arguments are always
// optional, and default to be a zero-length slice.
//
// Options whose type is bool cannot take a value. These options can only be set
// to true, and can only be included by being mentioned by name in os.Args.
// There are two syntaxes for doing this: the "short" form is for options that
// have a short name, and the "long" form is for options that have a long name.
// For example, if the option has short name "f" and long name "force", then
// these are equivalent:
//
//  // This is the "short" form
//  cmd -f
//
//  // This is the "long" form
//  cmd --force
//
// Options whose type is not bool and not a pointer must take a value. There are
// four syntaxes for setting the value for these options: the two "short" forms
// are for options that have a short name, and the two "long" forms are for
// options that have a long name. For example, if the option has short name "o"
// and long name "output", then these are all equivalent, and set the option to
// the string "json":
//
//  // This is the "short stuck" form
//  cmd -ojson
//
//  // This is the "short detached" form
//  cmd -o json
//
//  // This is the "long stuck" form
//  cmd --output=json
//
//  // This is the "long detached" form
//  cmd --output json
//
// Options whose type is a pointer may, but do not have to, take a value. To set
// the value, users must pass the value using either of the "stuck" forms of
// value-taking options immediately above. To avoid setting an explicit value,
// users must use either of the forms of the non-value-taking options further
// above.
//
// If a user sets a value for a pointer-typed option, then the corresponding
// field will be populated as a pointer to the parsed value. If the user
// specifies the option but does not give it a value, then the corresponding
// field will be populated as a pointer to the zero value of the underlying
// type. If the user does not specify the option at all, then the field will be
// populated as nil.
//
// When parsing sub-commands, the populated options for the parent command are
// set to the value of the field using the parent form of the "cli" tag. In
// other words: child commands can see the parsed options for their ancestor
// commands, by looking inside the value of the fields tagged with
// cli:"xxx,subcmd".
//
// If the user has specified the special, automatically-populated help option in
// their arguments, then Run will output the usage message of the appropriate
// command or sub-command. If the arguments in os.Args ultimately lead to a
// non-runnable command, then Run will similarly output the usage message of the
// relevant command.
//
// The usage message of a given command will contain the name of the command,
// its extended description, the name of its argument(s), and the names of its
// options and their usages.
//
// If Run calls one of the elements of funcs and that function retuns an error,
// then the error will be printed to os.Stderr and Run will call os.Exit(1).
//
// Man Page Generation
//
// If the UCARION_CLI_GENERATE_MAN environment variable is non-empty and the
// COMP_LINE and COMP_CWORD environment variables are empty, then Run will not
// parse os.Args and will instead generate man pages.
//
// Man pages will be generated in the directory specified by
// UCARION_CLI_GENERATE_MAN. Each command and sub-command within the command
// tree constructed from funcs will have its own man page.
//
// The man page of a given command will contain the name of the command, its
// description and extended description, the name of its argument(s), the names
// of its options and their extended usages.
//
// If Run encounters an I/O error while generating man pages, it panics.
//
// Bash or Zsh Completion
//
// If the COMP_LINE and COMP_CWORD environment variables are both non-empty,
// then Run will not run any of funcs, and will instead output a set of
// completions to stdout. In effect, this makes Run be its own Bash/Zsh
// completion script.
//
// When generating completions, Run will construct a command tree from funcs,
// and will parse arguments from COMP_LINE (separated by strings.Fields) up to
// the index COMP_CWORD (parsed by strconv.Atoi). Then, Run may call the
// autocompleter function of the relevant option or argument that is expected
// next in the input, if such an autocompleter is available.
//
// When Run calls an autocompleter, the method receiver will be a
// partially-populated config type. Autocompleter functions may rely on that
// partially-populated config type to inform completion generation; some
// applications may, for example, perform an authenticated request to some
// system if the user has specified credentials, and use the result of that
// request to return a set of suggestions.
//
// For guidance on how to usefully set up a Run-using application to have
// Bash/Zsh completions, see the README for cli, available online at:
//
// https://github.com/ucarion/cli
//
// That README also contains cookbook-style documentation on how to use Run; it
// is expected that for most users, the README will be more useful than these
// docs, which serve more as a description of the contract that Run upholds.
func Run(ctx context.Context, funcs ...interface{}) {
	// If we fail to build a tree from the user's given functions, then we
	// should panic. Panicking early makes an experience similar to a
	// compilation error, where the program fails very early and with a message
	// meant for the program's developer, not its end user.
	tree, err := cmdtree.New(funcs)
	if err != nil {
		panic(err)
	}

	// Try to see if we are being called as Bash autocompleter.
	completeLine := os.Getenv(envCompleteLine)
	completeArgc := os.Getenv(envCompleteArgc)

	if completeLine != "" && completeArgc != "" {
		// We are being invoked as a Bash autocompleter.
		//
		// If we encounter errors during this procedure, there isn't much we can
		// do -- there is no interface by which an autocompleter can give back
		// an error. The best we can do is be silent.
		args := strings.Fields(completeLine)
		argc, err := strconv.Atoi(completeArgc)
		if err != nil {
			return
		}

		// With the suggestions in hand, output each of them as a separate line
		// to stdout.
		for _, s := range autocompleter.Autocomplete(tree, args[:argc]) {
			fmt.Println(s)
		}

		return
	}

	// Try to see if we are being called by the user for the purposes of
	// generating the command tree's man pages.
	if dir := os.Getenv(envGenerateManDir); dir != "" {
		for name, contents := range cmdman.Man(tree, os.Args[0]) {
			filename := filepath.Join(dir, name)

			// Write the file with mode 0666. This is the default with
			// os.Create, and corresponds to a file that's universally
			// read/write-able, but not executable.
			if err := ioutil.WriteFile(filename, []byte(contents), 0666); err != nil {
				// There isn't much we can usefully do in this situation. When
				// we're being invoked in order to generate man pages, we are
				// being invoked from the CLI. So we can't return an error
				// through any normal Go mechanism.
				//
				// The important thing is that we exit with a useful error
				// message and a non-zero exit code, so that whatever shell or
				// script that's invoking us can see that man page generation
				// did not succeed.
				panic(err)
			}
		}

		return
	}

	// Run the args against the user's command tree.
	if err := exectree.Exec(ctx, tree, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
