package noderunscript

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		envRunScripts, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.DetectResult{},
				packit.Fail.WithMessage("environment variable $BP_NODE_RUN_SCRIPTS is not set")
		}

		_, err := os.Stat(filepath.Join(context.WorkingDir, "package.json"))
		if err != nil {
			return packit.DetectResult{},
				packit.Fail.WithMessage("file package.json does not exist")
		}

		packageJSONFile, err := ioutil.ReadFile(filepath.Join(context.WorkingDir, "package.json"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		var packageJSON struct {
			Name    string            `json:"name"`
			Scripts map[string]string `json:"scripts"`
		}

		err = json.Unmarshal([]byte(packageJSONFile), &packageJSON)
		if err != nil {
			return packit.DetectResult{}, err
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		for _, envScriptName := range envScriptNames {
			if _, exists := packageJSON.Scripts[envScriptName]; !exists {
				return packit.DetectResult{},
					fmt.Errorf("one of the scripts in $BP_NODE_RUN_SCRIPTS does not exist in package.json")
			}
		}

		lockName := "npm"
		if _, err := os.Stat(filepath.Join(context.WorkingDir, "yarn.lock")); err == nil {
			lockName = "yarn"
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{Name: "node"}, {Name: lockName},
				},
			},
		}, nil
	}
}
