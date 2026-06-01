package applicationcommands

import (
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	base := ApplicationCommand{
		Type:        CommandTypeChatInput,
		Name:        "lookup",
		Description: "Look up something",
		Options: []ApplicationCommandOption{
			{
				Type:        OptionTypeString,
				Name:        "query",
				Description: "Search text",
				Required:    true,
			},
		},
	}

	tests := []struct {
		name    string
		command ApplicationCommand
		wantErr string
	}{
		{
			name:    "valid chat input command",
			command: base,
		},
		{
			name: "required options must precede optional options",
			command: commandWithOptions(
				option("first", OptionTypeString, false),
				option("second", OptionTypeString, true),
			),
			wantErr: "required option",
		},
		{
			name: "choices and autocomplete are exclusive",
			command: commandWithOptions(ApplicationCommandOption{
				Type:         OptionTypeString,
				Name:         "query",
				Description:  "Search text",
				Autocomplete: true,
				Choices: []ApplicationCommandChoice{
					{Name: "alpha", Value: "alpha"},
				},
			}),
			wantErr: "choices and autocomplete",
		},
		{
			name: "subcommand group cannot contain leaf options",
			command: commandWithOptions(ApplicationCommandOption{
				Type:        OptionTypeSubCommandGroup,
				Name:        "admin",
				Description: "Admin actions",
				Options: []ApplicationCommandOption{
					option("reason", OptionTypeString, false),
				},
			}),
			wantErr: "can only contain subcommands",
		},
		{
			name: "context menu commands cannot have descriptions",
			command: ApplicationCommand{
				Type:        CommandTypeUser,
				Name:        "inspect",
				Description: "Not allowed",
			},
			wantErr: "context menu commands",
		},
		{
			name: "localization names must match command name rules",
			command: ApplicationCommand{
				Type:              CommandTypeChatInput,
				Name:              "lookup",
				Description:       "Look up something",
				NameLocalizations: map[string]string{"en-US": "Bad Name"},
			},
			wantErr: "name_localizations",
		},
		{
			name: "string min length cannot exceed max length",
			command: commandWithOptions(ApplicationCommandOption{
				Type:        OptionTypeString,
				Name:        "query",
				Description: "Search text",
				MinLength:   intPtr(5),
				MaxLength:   intPtr(2),
			}),
			wantErr: "min_length cannot exceed max_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateCommand() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateCommand() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func commandWithOptions(options ...ApplicationCommandOption) ApplicationCommand {
	return ApplicationCommand{
		Type:        CommandTypeChatInput,
		Name:        "lookup",
		Description: "Look up something",
		Options:     options,
	}
}

func option(name string, typ OptionType, required bool) ApplicationCommandOption {
	return ApplicationCommandOption{
		Type:        typ,
		Name:        name,
		Description: "Option description",
		Required:    required,
	}
}

func intPtr(value int) *int {
	return &value
}
