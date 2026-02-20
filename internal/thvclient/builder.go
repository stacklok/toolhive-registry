// Package thvclient provides a client for interacting with the ToolHive CLI
// to run MCP servers and discover their available tools.
package thvclient

import (
	"fmt"

	"github.com/stacklok/toolhive/pkg/registry/registry"

	"github.com/stacklok/toolhive-catalog/internal/serverjson"
)

// CommandBuilder helps build command line arguments for thv.
type CommandBuilder struct {
	args []string
}

// NewCommandBuilder creates a new command builder with the given base command.
func NewCommandBuilder(command string) *CommandBuilder {
	return &CommandBuilder{
		args: []string{command},
	}
}

// AddFlag adds a flag with a value. No-op if value is empty.
func (b *CommandBuilder) AddFlag(flag, value string) *CommandBuilder {
	if value != "" {
		b.args = append(b.args, flag, value)
	}
	return b
}

// AddBoolFlag adds a boolean flag when value is true.
func (b *CommandBuilder) AddBoolFlag(flag string, value bool) *CommandBuilder {
	if value {
		b.args = append(b.args, flag)
	}
	return b
}

// AddEnvVar adds an environment variable flag (-e NAME=value).
func (b *CommandBuilder) AddEnvVar(name, value string) *CommandBuilder {
	if value != "" {
		b.args = append(b.args, "-e", fmt.Sprintf("%s=%s", name, value))
	}
	return b
}

// AddPositional adds a positional argument.
func (b *CommandBuilder) AddPositional(value string) *CommandBuilder {
	b.args = append(b.args, value)
	return b
}

// Build returns the built command arguments.
func (b *CommandBuilder) Build() []string {
	return b.args
}

// BuildRunCommand builds thv run arguments from a ServerFile and extensions.
func BuildRunCommand(
	sf *serverjson.ServerFile,
	ext *registry.ServerExtensions,
	tempName, image string,
) []string {
	builder := NewCommandBuilder("run")
	builder.AddFlag("--name", tempName)

	// Transport from the package
	if len(sf.ServerJSON.Packages) > 0 {
		builder.AddFlag(
			"--transport",
			sf.ServerJSON.Packages[0].Transport.Type,
		)

		// Environment variables from package definition
		for _, ev := range sf.ServerJSON.Packages[0].EnvironmentVariables {
			if ev.Default != "" {
				builder.AddEnvVar(ev.Name, ev.Default)
				continue
			}
			if ev.IsRequired {
				builder.AddEnvVar(ev.Name, "placeholder")
				continue
			}
			if ev.IsSecret {
				builder.AddEnvVar(ev.Name, "placeholder")
			}
		}
	}

	// Permission profile from extensions
	if ext != nil && ext.Permissions != nil && ext.Permissions.Network != nil {
		builder.AddFlag("--permission-profile", "network")
	}

	// Image as positional argument
	builder.AddPositional(image)

	// Args from extensions (after "--" separator)
	if ext != nil && len(ext.Args) > 0 {
		builder.AddPositional("--")
		for _, a := range ext.Args {
			if a != "" {
				builder.AddPositional(a)
			}
		}
	}

	return builder.Build()
}
