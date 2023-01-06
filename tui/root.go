package tui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

type SunbeamConfig struct {
	Height      int
	Width       int
	FullScreen  bool
	AccentColor string
	CopyCommand string

	RootItems []app.RootItem `yaml:"rootItems"`
}

var Config SunbeamConfig = SunbeamConfig{
	Height:      0,
	Width:       0,
	AccentColor: "13",
	FullScreen:  true,
}

func init() {
	viper.AddConfigPath(app.Sunbeam.ConfigRoot)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sunbeam")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("accentColor", "13")
	viper.SetDefault("height", 0)

	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Printf("unable to decode config into struct, %v", err)
	}
}

type Container interface {
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type Model struct {
	width, height int
	exitCmd       *exec.Cmd
	exit          bool

	pages []Container

	hidden bool
}

func NewModel(rootPage Container) *Model {
	return &Model{pages: []Container{rootPage}}
}

func (m *Model) Init() tea.Cmd {
	return m.pages[0].Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.exit = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case OpenMsg:
		err := open.Run(msg.Target)
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit

	case CopyTextMsg:
		err := clipboard.WriteAll(msg.Text)
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case RunScriptMsg:
		extension, ok := app.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}

		if len(extension.Requirements) > 0 {
			for _, requirement := range extension.Requirements {
				if !requirement.Check() {
					container := NewDetail("Requirement not met")
					container.content = fmt.Sprintf("requirement %s not met.\nHomepage: %s", requirement.Which, requirement.HomePage)
					return m, NewPushCmd(container)
				}
			}
		}

		script, ok := extension.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s not found", msg.Script))
		}

		runner := NewScriptRunner(extension, script, msg.With)
		runner.OnSuccessCmd = msg.OnSuccessCmd()
		cmd := m.Push(runner)
		return m, cmd
	case ExecCommandMsg:
		command := exec.Command("sh", "-c", msg.Command)
		command.Dir = msg.Directory
		command.Env = os.Environ()
		command.Env = append(command.Env, msg.Env...)

		if msg.OnSuccess == "" {
			m.exitCmd = command
			m.exit = true
			return m, tea.Quit
		}

		m.hidden = true
		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			if err != nil {
				return showMsg{
					cmd: NewErrorCmd(err),
				}
			}

			return showMsg{
				cmd: msg.OnSuccessCmd(),
			}
		})
	case showMsg:
		m.hidden = false
		return m, msg.cmd
	case pushMsg:
		m.hidden = false
		cmd := m.Push(msg.container)
		return m, cmd
	case popMsg:
		if len(m.pages) == 1 {
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case error:
		log.Printf("error: %s", msg)
		detail := NewDetail("Error")
		detail.SetSize(m.width, m.pageHeight())
		detail.SetContent(msg.Error())
		m.pages[len(m.pages)-1] = detail

		return m, detail.Init()
	}

	// Update the current page
	var cmd tea.Cmd

	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)

	return m, cmd
}

func (m *Model) View() string {
	if m.hidden {
		return ""
	}

	var embedView string
	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		embedView = currentPage.View()
	} else {
		embedView = "No pages"
	}

	return embedView
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.width, m.pageHeight())
	}
}

func (m *Model) pageHeight() int {
	if Config.Height > 0 {
		return utils.Min(Config.Height, m.height)
	} else {
		return m.height
	}
}

type popMsg struct{}

type showMsg struct {
	cmd tea.Cmd
}

func PopCmd() tea.Msg {
	return popMsg{}
}

type pushMsg struct {
	container Container
}

func NewPushCmd(c Container) tea.Cmd {
	return func() tea.Msg {
		return pushMsg{c}
	}
}

func (m *Model) Push(page Container) tea.Cmd {
	page.SetSize(m.width, m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Model) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func toShellCommand(rootItem app.RootItem) string {
	args := []string{"sunbeam", "run", rootItem.Extension, rootItem.Script}
	for param, value := range rootItem.With {
		switch value := value.(type) {
		case string:
			value = shellescape.Quote(value)
			args = append(args, fmt.Sprintf("--%s=%s", param, value))
		case bool:
			if !value {
				continue
			}
			args = append(args, fmt.Sprintf("--%s", param))
		}
	}
	return strings.Join(args, " ")
}

func loadHistory(historyPath string) map[string]int64 {
	history := make(map[string]int64)
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return history
	}

	json.Unmarshal(data, &history)
	return history
}

func NewRootList(rootItems ...app.RootItem) Container {

	stateDir := path.Join(os.Getenv("HOME"), ".local", "state", "sunbeam")
	historyPath := path.Join(stateDir, "history.json")

	history := loadHistory(historyPath)

	list := NewList("Sunbeam")

	list.filter.Less = func(i, j FilterItem) bool {
		iValue, ok := history[i.ID()]
		if !ok {
			iValue = 0
		}
		jValue, ok := history[j.ID()]
		if !ok {
			jValue = 0
		}

		return iValue > jValue
	}

	listItems := make([]ListItem, 0)
	for _, rootItem := range rootItems {
		rootItem := rootItem
		extension, ok := app.Sunbeam.Extensions[rootItem.Extension]
		if !ok {
			fmt.Fprintln(os.Stderr, "extension not found:", rootItem.Extension)
			continue
		}
		with := make(map[string]app.ScriptInputWithValue)
		itemShellCommand := toShellCommand(rootItem)
		for key, value := range rootItem.With {
			with[key] = app.ScriptInputWithValue{Value: value}
		}
		listItems = append(listItems, ListItem{
			Id:       itemShellCommand,
			Title:    rootItem.Title,
			Subtitle: rootItem.Subtitle,
			Actions: []Action{
				{
					Title:    "Run Script",
					Shortcut: "enter",
					Cmd: func() tea.Msg {
						history[itemShellCommand] = time.Now().Unix()
						if _, err := os.Stat(stateDir); os.IsNotExist(err) {
							os.MkdirAll(stateDir, 0755)
						}

						data, _ := json.Marshal(history)
						os.WriteFile(historyPath, data, 0644)

						return RunScriptMsg{
							Extension: rootItem.Extension,
							Script:    rootItem.Script,
							With:      with,
						}
					},
				},
				{
					Title:    "Edit Script Manifest",
					Shortcut: "ctrl+e",
					Cmd:      NewEditCmd(extension.Url.Path),
				},
				{
					Title:    "Copy as Shell Command",
					Shortcut: "ctrl+y",
					Cmd:      NewCopyTextCmd(itemShellCommand),
				},
			},
		})
	}

	list.SetItems(listItems)

	return list
}

func Draw(model *Model) (err error) {
	// Log to a file
	if env := os.Getenv("SUNBEAM_LOG_FILE"); env != "" {
		f, err := tea.LogToFile(env, "debug")
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		defer f.Close()
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		logDir := path.Join(home, ".local", "state", "sunbeam")
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(home, ".local", "state", "sunbeam"), 0755)
			if err != nil {
				return err
			}
		}
		tea.LogToFile(path.Join(logDir, "sunbeam.log"), "")
	}

	var p *tea.Program
	if Config.FullScreen {
		p = tea.NewProgram(model, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(model)
	}

	m, err := p.Run()
	if err != nil {
		return err
	}

	model = m.(*Model)

	if exitCmd := model.exitCmd; exitCmd != nil {
		exitCmd.Stderr = os.Stderr
		exitCmd.Stdout = os.Stdout
		exitCmd.Stdin = os.Stdin

		exitCmd.Run()
	}

	return nil
}
