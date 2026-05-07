package model

type Plugin struct {
	Name       string `json:"name"`
	PluginName string `json:"plugin_name"`
	Type       string `json:"type"`
}

type ResolutionResult struct {
	Plugins []Plugin `json:"plugins"`
	Error   string   `json:"error,omitempty"`
}
