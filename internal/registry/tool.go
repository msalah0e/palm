package registry

// Tool represents an AI CLI tool in the registry.
type Tool struct {
	Name        string   `toml:"name"`
	DisplayName string   `toml:"display_name"`
	Description string   `toml:"description"`
	Category    string   `toml:"category"`
	Tags        []string `toml:"tags"`
	Homepage    string   `toml:"homepage"`
	Repo        string   `toml:"repo"`
	Install     Install  `toml:"install"`
	Keys        Keys     `toml:"keys"`
}

// Install defines how to install a tool via different backends.
type Install struct {
	Brew   string `toml:"brew"`
	Pip    string `toml:"pip"`
	Npm    string `toml:"npm"`
	Cargo  string `toml:"cargo"`
	Go     string `toml:"go"`
	Binary string `toml:"binary"`
	Script string `toml:"script"`
	Verify Verify `toml:"verify"`
}

// Verify defines how to check if a tool is installed.
type Verify struct {
	Command string `toml:"command"`
}

// Keys defines API key requirements for a tool.
type Keys struct {
	Required  []string `toml:"required"`
	Optional  []string `toml:"optional"`
	EnvPrefix string   `toml:"env_prefix"`
}

// InstallMethod returns the preferred install backend and package identifier.
func (t Tool) InstallMethod() (backend, pkg string) {
	switch {
	case t.Install.Brew != "":
		return "brew", t.Install.Brew
	case t.Install.Script != "":
		return "script", t.Install.Script
	case t.Install.Binary != "":
		return "binary", t.Install.Binary
	case t.Install.Pip != "":
		return "pip", t.Install.Pip
	case t.Install.Npm != "":
		return "npm", t.Install.Npm
	case t.Install.Cargo != "":
		return "cargo", t.Install.Cargo
	case t.Install.Go != "":
		return "go", t.Install.Go
	default:
		return "manual", t.Homepage
	}
}

// NeedsAPIKey returns true if the tool requires at least one API key.
func (t Tool) NeedsAPIKey() bool {
	return len(t.Keys.Required) > 0
}
