package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	history, err := tui.LoadHistory(path.Join(homeDir, ".local", "share", "sunbeam", "history.json"))
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to load history: %w", err)
	}

	var config tui.Config
	configPath := path.Join(homeDir, ".config", "sunbeam", "config.yml")
	if _, err := os.Stat(configPath); err == nil {
		bytes, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(bytes, &config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(extensionRoot, 0755); err != nil {
			return err
		}
	}

	extensions, err := app.LoadExtensions(extensionRoot)
	if err != nil {
		return err
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Version:      version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootList := tui.NewRootList(history, extensions...)
			model := tui.NewModel(rootList)
			return tui.Draw(model)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		Title: "Core Commands",
		ID:    "core",
	}, &cobra.Group{
		Title: "Extension Commands",
		ID:    "extension",
	})

	// Core Commands
	rootCmd.AddCommand(NewCmdExtension(extensionRoot, extensions))
	rootCmd.AddCommand(NewCmdServe(extensions))
	rootCmd.AddCommand(NewCmdValidate())
	rootCmd.AddCommand(NewCmdRun(&config))

	for _, extension := range extensions {
		rootCmd.AddCommand(NewExtensionCommand(extension, history, &config))
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension app.Extension, history *tui.History, config *tui.Config) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     extension.Name(),
		GroupID: "extension",
		Short:   extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootList := tui.NewRootList(history, extension)
			model := tui.NewModel(rootList)
			err = tui.Draw(model)
			if err != nil {
				return fmt.Errorf("could not run extension: %w", err)
			}

			return nil
		},
	}

	for _, command := range extension.Commands {
		subcommand := NewExtensionSubCommand(extension, command)
		extensionCmd.AddCommand(subcommand)
	}

	return extensionCmd
}

func NewExtensionSubCommand(extension app.Extension, command app.Command) *cobra.Command {
	cmd := cobra.Command{
		Use:   command.Name,
		Short: command.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			with := make(map[string]app.Arg)
			for _, param := range command.Params {
				if !cmd.Flags().Changed(param.Name) {
					continue
				}

				switch param.Type {
				case "boolean":
					value, err := cmd.Flags().GetBool(param.Name)
					if err != nil {
						return err
					}
					with[param.Name] = app.Arg{Value: value}
				default:
					value, err := cmd.Flags().GetString(param.Name)
					if err != nil {
						return err
					}
					with[param.Name] = app.Arg{Value: value}
				}

			}
			runner := tui.NewCommandRunner(
				extension,
				command,
				with,
			)

			model := tui.NewModel(runner)

			err = tui.Draw(model)
			if err != nil {
				return fmt.Errorf("could not run script: %w", err)
			}
			return nil
		},
	}

	for _, param := range command.Params {
		switch param.Type {
		case "boolean":
			if param.Default != nil {
				defaultValue := param.Default.(bool)
				cmd.Flags().Bool(param.Name, defaultValue, param.Description)
			} else {
				cmd.Flags().Bool(param.Name, false, param.Description)
			}
		default:
			if param.Default != nil {
				defaultValue := param.Default.(string)
				cmd.Flags().String(param.Name, defaultValue, param.Description)
			} else {
				cmd.Flags().String(param.Name, "", param.Description)
			}
		}

		if param.Type == "file" {
			cmd.MarkFlagFilename(param.Name)
		}

		if param.Type == "directory" {
			cmd.MarkFlagDirname(param.Name)
		}
	}

	return &cmd

}
