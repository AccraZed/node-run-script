package noderunscript

import (
	"encoding/json"
	"os"
	"path/filepath"
)

//go:generate faux --interface PackageInterface -o fakes/package_interface.go
type PackageInterface interface {
	GetPackageScripts(path string) (map[string]string, error)
	GetPackageManager(path string) string
}

type ScriptManager struct{}

func CreateScriptManager() *ScriptManager {
	return &ScriptManager{}
}

func (s *ScriptManager) GetPackageScripts(path string) (map[string]string, error) {
	packageJSONFile, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return nil, err
	}

	var packageJSON struct {
		Scripts map[string]string `json:"scripts"`
	}

	err = json.Unmarshal([]byte(packageJSONFile), &packageJSON)
	if err != nil {
		return nil, err
	}

	return packageJSON.Scripts, nil
}

func (s *ScriptManager) GetPackageManager(path string) string {
	_, err := os.Stat(filepath.Join(path, "yarn.lock"))
	if err == nil {
		return "yarn"
	}
	return "npm"
}
