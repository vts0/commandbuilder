package commandbuilder

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Command represents a command with its arguments, environment variables, redirections, and chaining logic.
type Command struct {
	name         string
	subcommands  []string
	args         []argument
	useSudo      bool
	next         *Command
	operator     string
	redirections []string
	background   bool
	env          map[string]string
	stderrRedir  string
	mergeStdErr  bool
	stdin        io.Reader
	group        bool
	tempFiles    []tempFile
}

// argument represents a command argument with additional metadata.
type argument struct {
	value    string
	quoted   bool
	expand   bool
	isPath   bool
	isGlob   bool
	keyValue bool
	key      string
}

// tempFile represents a temporary file to be created for the command.
type tempFile struct {
	content string
	mode    os.FileMode
}

// CommandBuilder provides a fluent interface for building commands.
type CommandBuilder struct {
	command *Command
}

// New creates a new CommandBuilder with the specified command name.
func New(name string) *CommandBuilder {
	return &CommandBuilder{
		command: &Command{
			name:         name,
			args:         make([]argument, 0),
			env:          make(map[string]string),
			subcommands:  make([]string, 0),
			tempFiles:    make([]tempFile, 0),
			redirections: make([]string, 0),
		},
	}
}

// WithArgument adds a plain argument to the command.
func (b *CommandBuilder) WithArgument(value string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: value})
	return b
}

// WithOption adds the option and its value (if a value is passed).
func (b *CommandBuilder) WithOption(option string, values ...string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: option})
	if len(values) > 0 {
		b.command.args = append(b.command.args, argument{value: values[0]})
	}
	return b
}

// WithSubcommand adds a subcommand (e.g., "commit" in "git commit").
func (b *CommandBuilder) WithSubcommand(sub string) *CommandBuilder {
	b.command.subcommands = append(b.command.subcommands, sub)
	return b
}

// WithQuotedArgument adds an argument that will be wrapped in quotes.
func (b *CommandBuilder) WithQuotedArgument(value string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: value, quoted: true})
	return b
}

// WithKeyValueArgument adds a key-value argument (e.g., "--key=value").
func (b *CommandBuilder) WithKeyValueArgument(key, value string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{
		keyValue: true,
		key:      key,
		value:    value,
	})
	return b
}

// WithGlobArgument adds a glob pattern argument (e.g., "*.txt").
func (b *CommandBuilder) WithGlobArgument(pattern string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: pattern, isGlob: true})
	return b
}

// WithPathArgument adds a path argument (e.g., "/path/with spaces").
func (b *CommandBuilder) WithPathArgument(path string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: path, isPath: true})
	return b
}

// WithVariable adds a variable argument (e.g., "$HOME").
func (b *CommandBuilder) WithVariable(name string) *CommandBuilder {
	b.command.args = append(b.command.args, argument{value: name, expand: true})
	return b
}

// WithSudo enables sudo for the command.
func (b *CommandBuilder) WithSudo() *CommandBuilder {
	b.command.useSudo = true
	return b
}

// WithEnv sets an environment variable for the command.
func (b *CommandBuilder) WithEnv(key, value string) *CommandBuilder {
	b.command.env[key] = value
	return b
}

// RedirectToDevNull redirects the command's output to /dev/null (overwrite).
func (b *CommandBuilder) RedirectToDevNull() *CommandBuilder {
	b.command.redirections = append(b.command.redirections, "> /dev/null")
	return b
}

// RedirectToFile redirects the command's output to a file (overwrite).
func (b *CommandBuilder) RedirectToFile(filename string) *CommandBuilder {
	b.command.redirections = append(b.command.redirections, fmt.Sprintf("> %s", filename))
	return b
}

// AppendToFile redirects the command's output to a file (append).
func (b *CommandBuilder) AppendToFile(filename string) *CommandBuilder {
	b.command.redirections = append(b.command.redirections, fmt.Sprintf(">> %s", filename))
	return b
}

// RedirectFromFile redirects the command's input from a file.
func (b *CommandBuilder) RedirectFromFile(filename string) *CommandBuilder {
	b.command.redirections = append(b.command.redirections, fmt.Sprintf("< %s", filename))
	return b
}

// RedirectStderrToFile redirects stderr to a file.
func (b *CommandBuilder) RedirectStderrToFile(filename string) *CommandBuilder {
	b.command.stderrRedir = fmt.Sprintf("2> %s", filename)
	return b
}

// MergeStdoutAndStderr merges stderr with stdout.
func (b *CommandBuilder) MergeStdoutAndStderr() *CommandBuilder {
	b.command.mergeStdErr = true
	return b
}

// WithStdin sets the command's standard input.
func (b *CommandBuilder) WithStdin(reader io.Reader) *CommandBuilder {
	b.command.stdin = reader
	return b
}

// WithTempFile creates a temporary file for the command.
func (b *CommandBuilder) WithTempFile(content string, mode os.FileMode) *CommandBuilder {
	b.command.tempFiles = append(b.command.tempFiles, tempFile{
		content: content,
		mode:    mode,
	})
	return b
}

// Grouped groups the command and its chained commands in a subshell.
func (b *CommandBuilder) Grouped() *CommandBuilder {
	b.command.group = true
	return b
}

// Background runs the command in the background.
func (b *CommandBuilder) Background() *CommandBuilder {
	b.command.background = true
	return b
}

// PipeTo chains the current command to the next command using a pipe ("|").
func (b *CommandBuilder) PipeTo(next *CommandBuilder) *CommandBuilder {
	return b.chain(next, "|")
}

// And chains the current command to the next command using a logical AND ("&&").
func (b *CommandBuilder) And(next *CommandBuilder) *CommandBuilder {
	return b.chain(next, "&&")
}

// Or chains the current command to the next command using a logical OR ("||").
func (b *CommandBuilder) Or(next *CommandBuilder) *CommandBuilder {
	return b.chain(next, "||")
}

// chain is a helper method for chaining commands with an operator.
func (b *CommandBuilder) chain(next *CommandBuilder, op string) *CommandBuilder {
	b.command.operator = op
	b.command.next = next.command
	return b
}

// Build constructs the final command string.
func (b *CommandBuilder) Build() string {
	var parts []string
	cmd := b.command

	for cmd != nil {
		var segmentParts []string

		// Environment variables
		if len(cmd.env) > 0 {
			envParts := make([]string, 0, len(cmd.env))
			for k, v := range cmd.env {
				envParts = append(envParts, fmt.Sprintf("%s=%s", k, shellEscape(v)))
			}
			segmentParts = append(segmentParts, strings.Join(envParts, " "))
		}

		// Command name and subcommands
		cmdParts := []string{shellEscape(cmd.name)}
		for _, sub := range cmd.subcommands {
			cmdParts = append(cmdParts, shellEscape(sub))
		}

		// Arguments processing
		for _, arg := range cmd.args {
			var part string
			switch {
			case arg.keyValue:
				part = fmt.Sprintf("%s=%s", arg.key, processArgument(arg.value, arg))
			case arg.quoted:
				part = fmt.Sprintf(`"%s"`, processArgument(arg.value, arg))
			case arg.expand:
				part = fmt.Sprintf("$%s", arg.value)
			case arg.isGlob:
				part = arg.value // Globs should not be escaped
			case arg.isPath:
				part = shellEscapePath(arg.value)
			default:
				part = processArgument(arg.value, arg)
			}
			cmdParts = append(cmdParts, part)
		}

		// Sudo
		if cmd.useSudo {
			cmdParts = append([]string{"sudo"}, cmdParts...)
		}

		// Grouping
		if cmd.group {
			segmentParts = append(segmentParts, "("+strings.Join(cmdParts, " ")+")")
		} else {
			segmentParts = append(segmentParts, strings.Join(cmdParts, " "))
		}

		// Redirections
		if len(cmd.redirections) > 0 {
			segmentParts = append(segmentParts, strings.Join(cmd.redirections, " "))
		}

		// stderr redirection
		if cmd.stderrRedir != "" {
			segmentParts = append(segmentParts, cmd.stderrRedir)
		}

		// Merge stderr with stdout
		if cmd.mergeStdErr {
			segmentParts = append(segmentParts, "2>&1")
		}

		// Background execution
		if cmd.background {
			segmentParts = append(segmentParts, "&")
		}

		// Join segment parts
		segment := strings.TrimSpace(strings.Join(segmentParts, " "))
		parts = append(parts, segment)

		// Process next command
		if cmd.next != nil {
			parts = append(parts, cmd.operator)
		}
		cmd = cmd.next
	}

	return strings.Join(parts, " ")
}

// Helper functions
func processArgument(value string, arg argument) string {
	if arg.isPath {
		return shellEscapePath(value)
	}
	return shellEscape(value)
}

func shellEscape(s string) string {
	if strings.ContainsAny(s, " \t\n\"'$&;|<>`") {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "'\"'\"'"))
	}
	return s
}

func shellEscapePath(path string) string {
	// Special handling for paths with spaces
	return fmt.Sprintf("%q", path)
}
