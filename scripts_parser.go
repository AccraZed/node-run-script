package noderunscript

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func getPackageScripts(workingDir string) (map[string]string, error) {
	packageJSONFile, err := ioutil.ReadFile(filepath.Join(workingDir, "package.json"))
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
