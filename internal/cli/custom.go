package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Manifest.Title,
		Long:    extension.Manifest.Description,
		Args:    cobra.NoArgs,
		GroupID: CommandGroupExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			var inputBytes []byte
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				inputBytes = b
			}

			if len(inputBytes) == 0 {
				if !isatty.IsTerminal(os.Stdout.Fd()) {
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					encoder.SetEscapeHTML(false)

					return encoder.Encode(extension.Manifest)
				}

				if len(extension.Manifest.Root) == 0 {
					return cmd.Usage()
				}

				history, err := history.Load(history.Path)
				if err != nil {
					return err
				}

				rootList := tui.NewRootList(extension.Manifest.Title, history, func() ([]sunbeam.ListItem, error) {
					return extension.RootItems(), nil
				})

				return tui.Draw(rootList)
			}

			var input sunbeam.Payload
			if err := json.Unmarshal(inputBytes, &input); err != nil {
				return err
			}

			return runExtension(extension, input)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	commands := extension.Manifest.Commands
	sort.Slice(extension.Manifest.Commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	for _, command := range commands {
		cmd := NewSubCmdCustom(alias, extension, command)
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func NewSubCmdCustom(alias string, extension extensions.Extension, command sunbeam.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Name,
		Short: command.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := make(map[string]any)

			for _, param := range command.Params {
				if !cmd.Flags().Changed(param.Name) {
					continue
				}

				switch param.Type {
				case sunbeam.InputString:
					value, err := cmd.Flags().GetString(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.InputBoolean:
					value, err := cmd.Flags().GetBool(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.InputNumber:
					value, err := cmd.Flags().GetInt(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				}
			}

			input := sunbeam.Payload{
				Command: command.Name,
				Params:  params,
			}

			if !isatty.IsTerminal(os.Stdin.Fd()) {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				input.Query = string(bytes.Trim(stdin, "\n"))
			}

			return runExtension(extension, input)
		},
	}

	for _, input := range command.Params {
		switch input.Type {
		case sunbeam.InputString:
			cmd.Flags().String(input.Name, "", input.Title)
		case sunbeam.InputBoolean:
			cmd.Flags().Bool(input.Name, false, input.Title)
		case sunbeam.InputNumber:
			cmd.Flags().Int(input.Name, 0, input.Title)
		}

		if !input.Optional {
			_ = cmd.MarkFlagRequired(input.Name)
		}
	}

	return cmd
}

func runExtension(extension extensions.Extension, input sunbeam.Payload) error {
	command, ok := extension.Command(input.Command)
	if !ok {
		return fmt.Errorf("command %s not found", input.Command)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd, err := extension.Cmd(input)
		if err != nil {
			return err
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	switch command.Mode {
	case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
		runner := tui.NewRunner(extension, input)
		return tui.Draw(runner)
	case sunbeam.CommandModeSilent:
		return extension.Run(input)
	default:
		return fmt.Errorf("unknown command mode: %s", command.Mode)
	}
}
