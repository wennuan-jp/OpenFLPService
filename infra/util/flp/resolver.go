package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type Plugin struct {
	Name       string `json:"name"`
	PluginName string `json:"plugin_name"`
	Type       string `json:"type"`
}

type ResolutionResult struct {
	Plugins []Plugin `json:"plugins"`
	Error   string   `json:"error,omitempty"`
}

// ResolveFLP calls the python script to parse the FLP file
func ResolveFLP(filePath string) (*ResolutionResult, error) {
	// Assuming python3 is in the path and flp_resolver.py is in the same directory or relative to project root
	cmd := exec.Command("python3", "infra/flp_resolver.py", filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("python script failed: %v, stderr: %s", err, stderr.String())
	}

	var result ResolutionResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse json output: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("resolver error: %s", result.Error)
	}

	return &result, nil
}
