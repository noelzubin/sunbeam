package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/acarl005/stripansi"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Runner struct {
	embed         Page
	width, height int
	cancel        context.CancelFunc

	extension extensions.Extension
	command   types.CommandSpec
	input     types.CommandInput
}

func NewRunner(extension extensions.Extension, input types.CommandInput) *Runner {
	var embed Page
	command, ok := extension.Command(input.Command)
	if ok {
		switch command.Mode {
		case types.CommandModeList:
			list := NewList()
			if input.Query != "" {
				list.SetQuery(input.Query)
			}

			embed = list
		case types.CommandModeDetail:
			embed = NewDetail("")
		default:
			embed = NewErrorPage(fmt.Errorf("invalid view type"))
		}
	} else {
		embed = NewErrorPage(fmt.Errorf("command %s not found", input.Command))
	}

	return &Runner{
		embed:     embed,
		extension: extension,
		command:   command,
		input:     input,
	}
}

func (c *Runner) SetIsLoading(isLoading bool) tea.Cmd {
	switch page := c.embed.(type) {
	case *Detail:
		return page.SetIsLoading(isLoading)
	case *List:
		return page.SetIsLoading(isLoading)
	}

	return nil
}

func (c *Runner) Init() tea.Cmd {
	return tea.Batch(c.Reload(), c.embed.Init())
}

func (c *Runner) Focus() tea.Cmd {
	if c.embed == nil {
		return nil
	}
	termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Title))
	return c.embed.Focus()
}

func (c *Runner) Blur() tea.Cmd {
	c.cancel()
	return nil
}

func (c *Runner) SetSize(w int, h int) {
	c.width = w
	c.height = h

	c.embed.SetSize(w, h)
}

func (c *Runner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e":
			if c.extension.Type != extensions.ExtensionTypeLocal {
				break
			}
			editor := utils.FindEditor()
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, c.extension.Entrypoint))
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				return types.Action{
					Type: types.ActionTypeReload,
				}
			})
		case "esc":
			if c.embed != nil {
				break
			}
			return c, PopPageCmd
		case "ctrl+r":
			return c, c.Reload()
		}

	case Page:
		c.embed = msg
		c.embed.SetSize(c.width, c.height)
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case types.Action:
		switch msg.Type {
		case types.ActionTypeRun:
			command, ok := c.extension.Command(msg.Command)
			if !ok {
				c.embed = NewErrorPage(fmt.Errorf("command %s not found", msg.Command))
				c.embed.SetSize(c.width, c.height)
				return c, c.embed.Init()
			}

			input := types.CommandInput{
				Command:     msg.Command,
				Params:      msg.Params,
				Preferences: c.input.Preferences,
			}

			switch command.Mode {
			case types.CommandModeList, types.CommandModeDetail:
				runner := NewRunner(c.extension, input)

				return c, PushPageCmd(runner)
			case types.CommandModeSilent:
				return c, func() tea.Msg {
					_, err := c.extension.Output(input)

					if err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					if msg.Reload {
						return types.Action{
							Type: types.ActionTypeReload,
						}
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}
			case types.CommandModeTTY:
				cmd, err := c.extension.Cmd(input)

				if err != nil {
					c.embed = NewErrorPage(err)
					c.embed.SetSize(c.width, c.height)
					return c, c.embed.Init()
				}

				return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					if msg.Reload {
						return types.Action{
							Type: types.ActionTypeReload,
						}
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				})
			}
		case types.ActionTypeEdit:
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", utils.FindEditor(), msg.Target))
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Reload {
					return c.Reload()
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			})
		case types.ActionTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Text); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.ActionTypeOpen:
			command := msg
			return c, func() tea.Msg {
				if err := utils.OpenWith(command.Target, command.App); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.ActionTypeReload:
			if c.input.Params == nil {
				c.input.Params = make(map[string]any)
			}

			for k, v := range msg.Params {
				c.input.Params[k] = v
			}

			return c, c.Reload()
		case types.ActionTypeExit:
			return c, ExitCmd
		}
	case error:
		c.embed = NewErrorPage(msg)
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	}

	var cmd tea.Cmd
	c.embed, cmd = c.embed.Update(msg)
	return c, cmd
}

func (c *Runner) View() string {
	return c.embed.View()
}

func (c *Runner) Reload() tea.Cmd {
	return tea.Sequence(c.SetIsLoading(true), func() tea.Msg {
		if c.cancel != nil {
			c.cancel()
		}

		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		defer cancel()

		cmd, err := c.extension.CmdContext(ctx, c.input)
		if err != nil {
			return err
		}

		output, err := cmd.Output()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
			}

			return err
		}

		switch c.command.Mode {
		case types.CommandModeDetail:
			if err := schemas.ValidateDetail(output); err != nil {
				return err
			}

			var detail types.Detail
			if err := json.Unmarshal(output, &detail); err != nil {
				return err
			}

			page := NewDetail(detail.Text, detail.Actions...)
			if detail.Format != "" {
				page.Format = detail.Format
			}

			return page
		case types.CommandModeList:
			if err := schemas.ValidateList(output); err != nil {
				return err
			}

			var list types.List
			if err := json.Unmarshal(output, &list); err != nil {
				return err
			}

			page := NewList(list.Items...)
			if list.EmptyText != "" {
				page.SetEmptyText(list.EmptyText)
			}
			if list.Actions != nil {
				page.SetActions(list.Actions...)
			}

			if list.Dynamic {
				page.OnQueryChange = func(query string) tea.Cmd {
					c.input.Query = query
					return c.Reload()
				}
			}

			page.SetSize(c.width, c.height)
			if embed, ok := c.embed.(*List); ok {
				page.SetQuery(embed.Query())
			}

			return page
		default:
			return fmt.Errorf("invalid view type")
		}
	})
}
