# CommandBuilder

CommandBuilder is a Go package that provides a fluent interface for constructing and managing shell commands with arguments, environment variables, redirections, and chaining logic.

## Features
- Fluent interface for building commands
- Support for subcommands and arguments
- Environment variable handling
- Input/output redirection
- Command chaining with `|`, `&&`, `||`
- Background execution
- Grouping commands in subshells

## Installation
To install the package, run:
```sh
go get github.com/vts0/commandbuilder
```

## Usage

### Creating a Simple Command
```go
cmd := commandbuilder.New("ls").WithArgument("-la").Build()
fmt.Println(cmd) // Output: ls -la
```

### Using Subcommands
```go
cmd := commandbuilder.New("git").WithSubcommand("commit").WithOption("-m", "Initial commit").Build()
fmt.Println(cmd) // Output: git commit -m "Initial commit"
```

### Adding Environment Variables
```go
cmd := commandbuilder.New("echo").WithEnv("GREETING", "Hello").WithArgument("$GREETING").Build()
fmt.Println(cmd) // Output: GREETING=Hello echo $GREETING
```

### Redirecting Output
```go
cmd := commandbuilder.New("echo").WithArgument("Hello").RedirectToFile("output.txt").Build()
fmt.Println(cmd) // Output: echo Hello > output.txt
```

### Chaining Commands
```go
cmd := commandbuilder.New("echo").WithArgument("Hello").PipeTo(commandbuilder.New("grep").WithArgument("H")).Build()
fmt.Println(cmd) // Output: echo Hello | grep H
```

### Running in the Background
```go
cmd := commandbuilder.New("sleep").WithArgument("10").Background().Build()
fmt.Println(cmd) // Output: sleep 10 &
```

## API Reference

### CommandBuilder Methods
- `New(name string) *CommandBuilder` - Creates a new CommandBuilder.
- `WithArgument(value string) *CommandBuilder` - Adds an argument.
- `WithOption(option string, values ...string) *CommandBuilder` - Adds an option with optional values.
- `WithSubcommand(sub string) *CommandBuilder` - Adds a subcommand.
- `WithEnv(key, value string) *CommandBuilder` - Sets an environment variable.
- `RedirectToFile(filename string) *CommandBuilder` - Redirects output to a file.
- `PipeTo(next *CommandBuilder) *CommandBuilder` - Chains commands with a pipe (`|`).
- `And(next *CommandBuilder) *CommandBuilder` - Chains commands with `&&`.
- `Or(next *CommandBuilder) *CommandBuilder` - Chains commands with `||`.
- `Background() *CommandBuilder` - Runs the command in the background.
- `Build() string` - Constructs the final command string.

## License
This package is released under the MIT License.

## Contributing
Feel free to open issues or submit pull requests to improve this package.

## Author
[Your Name](https://github.com/vts0)

