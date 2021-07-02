package noderunscript

import (
	"errors"
	"fmt"
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
				packit.Fail.WithMessage("expected value from $BP_NODE_RUN_SCRIPTS to be set")
		}

		projectDir := context.WorkingDir
		bpNodeProjectPath, exists := os.LookupEnv("BP_NODE_PROJECT_PATH")
		if exists {
			projectDir = filepath.Join(context.WorkingDir, bpNodeProjectPath)
		}

		projectDir = filepath.Clean(projectDir)
		_, err := os.Stat(projectDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return packit.DetectResult{},
					fmt.Errorf("expected value from $BP_NODE_PROJECT_PATH [%s] to be an existing directory", projectDir)
			}
			return packit.DetectResult{}, err
		}

		_, err = os.Stat(filepath.Join(projectDir, "package.json"))
		if err != nil {
			return packit.DetectResult{},
				packit.Fail.WithMessage("expected file package.json to exist")
		}

		packageScripts, err := getPackageScripts(projectDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		for _, envScriptName := range envScriptNames {
			if _, exists := packageScripts[envScriptName]; !exists {
				return packit.DetectResult{},
					fmt.Errorf("expected a script from $BP_NODE_RUN_SCRIPTS to exist in package.json")
			}
		}

		lockName := getPackageManager(projectDir)

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{Name: "node"}, {Name: lockName},
				},
			},
		}, nil
	}
}
