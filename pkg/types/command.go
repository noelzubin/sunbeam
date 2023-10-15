package types

type CommandRef struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

// TODO: move distinct types to their own structs
type Command struct {
	Type CommandType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    Application `json:"app,omitempty"`
	Target string      `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type CommandType string

const (
	CommandTypeRun    CommandType = "run"
	CommandTypeOpen   CommandType = "open"
	CommandTypeCopy   CommandType = "copy"
	CommandTypeReload CommandType = "reload"
	CommandTypeExit   CommandType = "exit"
	CommandTypePop    CommandType = "pop"
)

type Application struct {
	Windows string `json:"windows,omitempty"`
	Mac     string `json:"mac,omitempty"`
	Linux   string `json:"linux,omitempty"`
}

type CommandInput struct {
	Params   map[string]any `json:"params"`
	FormData map[string]any `json:"formData,omitempty"`
	Query    string         `json:"query,omitempty"`
	WorkDir  string         `json:"workDir,omitempty"`
}
