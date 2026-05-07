package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"openflp.com/model"
)

type flpResolver struct {
	pythonScriptPath string
}

func NewFLPResolver(scriptPath string) *flpResolver {
	// If the path is relative, we try to resolve it from the project root
	if !filepath.IsAbs(scriptPath) {
		absPath, err := findFileInProjectRoot(scriptPath)
		if err == nil {
			scriptPath = absPath
		}
	}

	return &flpResolver{
		pythonScriptPath: scriptPath,
	}
}

func (r *flpResolver) Resolve(filePath string) (*model.ResolutionResult, error) {
	cmd := exec.Command("python3", r.pythonScriptPath, filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("python script failed: %v, stderr: %s", err, stderr.String())
	}

	var result model.ResolutionResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse json output: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("resolver error: %s", result.Error)
	}

	return &result, nil
}

// findFileInProjectRoot tries to find a file starting from the current directory and moving up until it finds go.mod
func findFileInProjectRoot(targetPath string) (string, error) {
	curr, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		potentialPath := filepath.Join(curr, targetPath)
		if _, err := os.Stat(potentialPath); err == nil {
			return potentialPath, nil
		}

		// Look for go.mod to stop
		if _, err := os.Stat(filepath.Join(curr, "go.mod")); err == nil {
			// Even if go.mod is here, we already checked the targetPath above.
			// If it's not here, it's not in the root either.
			break
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", fmt.Errorf("could not find %s in project root", targetPath)
}
