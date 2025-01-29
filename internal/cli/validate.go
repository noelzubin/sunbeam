package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		GroupID: CommandGroupCore,
		Short:   "Validate a Sunbeam schema",
	}

	cmd.AddCommand(NewCmdValidateList())
	cmd.AddCommand(NewCmdValidateDetail())
	cmd.AddCommand(NewCmdValidateManifest())

	return cmd
}

func NewCmdValidateList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"filter", "generator"},
		Short:   "Validate a list",
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidateList(input); err != nil {
				return fmt.Errorf("list is invalid: %s", err)
			}

			fmt.Println("✅ Valid!")
			return nil
		},
	}
}

func NewCmdValidateDetail() *cobra.Command {
	return &cobra.Command{
		Use:   "detail",
		Short: "Validate a detail",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidateDetail(input); err != nil {
				return fmt.Errorf("detail is invalid: %s", err)
			}

			fmt.Println("✅ List is valid!")
			return nil
		},
	}

}

func NewCmdValidateManifest() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Validate a manifest",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidateManifest(input); err != nil {
				return fmt.Errorf("manifest is invalid: %s", err)
			}

			fmt.Println("✅ Manifest is valid!")
			return nil
		},
	}
}
