package toolhive

import (
	"fmt"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

// CommandBuilder helps build command line arguments for thv
type CommandBuilder struct {
	args []string
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder(command string) *CommandBuilder {
	return &CommandBuilder{
		args: []string{command},
	}
}

// AddFlag adds a flag with a value
func (b *CommandBuilder) AddFlag(flag, value string) *CommandBuilder {
	if value != "" {
		b.args = append(b.args, flag, value)
	}
	return b
}

// AddBoolFlag adds a boolean flag
func (b *CommandBuilder) AddBoolFlag(flag string, value bool) *CommandBuilder {
	if value {
		b.args = append(b.args, flag)
	}
	return b
}

// AddEnvVar adds an environment variable
func (b *CommandBuilder) AddEnvVar(name, value string) *CommandBuilder {
	if value != "" {
		b.args = append(b.args, "-e", fmt.Sprintf("%s=%s", name, value))
	}
	return b
}

// AddPositional adds a positional argument
func (b *CommandBuilder) AddPositional(value string) *CommandBuilder {
	b.args = append(b.args, value)
	return b
}

// Build returns the built command arguments
func (b *CommandBuilder) Build() []string {
	return b.args
}

// BuildRunCommand builds the thv run command arguments from a spec
func BuildRunCommand(spec *types.RegistryEntry, tempName, image string) []string {
	builder := NewCommandBuilder("run")
	builder.AddFlag("--name", tempName)

	if spec.ImageMetadata != nil {
		// Add transport
		builder.AddFlag("--transport", spec.ImageMetadata.Transport)

		// Add environment variables
		if spec.ImageMetadata.EnvVars != nil {
			for _, envVar := range spec.ImageMetadata.EnvVars {
				if envVar.Secret {
					// For secrets, use placeholder values if required
					if envVar.Required {
						builder.AddEnvVar(envVar.Name, "placeholder")
					}
				} else if envVar.Default != "" {
					builder.AddEnvVar(envVar.Name, envVar.Default)
				}
			}
		}

		// Add permission profile
		if spec.Permissions != nil && spec.Permissions.Network != nil {
			builder.AddFlag("--permission-profile", "network")
		}
	}

	// Add the image as the last positional argument
	builder.AddPositional(image)

	return builder.Build()
}
