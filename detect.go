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

type packageJSON struct {
	Name    string            `json:"name"`
	Scripts map[string]string `json:"scripts"`
}

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		envRunScripts, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.DetectResult{},
				packit.Fail.WithMessage("environment variable $BP_NODE_RUN_SCRIPTS is not set")
		}

		if _, err := os.Stat(filepath.Join(context.WorkingDir, "package.json")); err != nil {
			return packit.DetectResult{},
				packit.Fail.WithMessage("file package.json does not exist")
		}

		jsonFile, err := ioutil.ReadFile(filepath.Join(context.WorkingDir, "package.json"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		var UnmarshalledJSON packageJSON
		err = json.Unmarshal([]byte(jsonFile), &UnmarshalledJSON)
		if err != nil {
			return packit.DetectResult{}, err
		}

		for _, envScriptName := range envScriptNames {
			if _, exists := UnmarshalledJSON.Scripts[envScriptName]; !exists {
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
