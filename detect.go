package noderunscript

import (
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
				packit.Fail.WithMessage("environment variable $BP_NODE_RUN_SCRIPTS is not set")
		}

		_, err := os.Stat(filepath.Join(context.WorkingDir, "package.json"))
		if err != nil {
			return packit.DetectResult{},
				packit.Fail.WithMessage("file package.json does not exist")
		}

		packageScripts, err := getPackageScripts(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		for _, envScriptName := range envScriptNames {
			if _, exists := packageScripts[envScriptName]; !exists {
				return packit.DetectResult{},
					fmt.Errorf("one of the scripts in $BP_NODE_RUN_SCRIPTS does not exist in package.json")
			}
		}

		lockName := getPackageManager(context.WorkingDir)

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{Name: "node"}, {Name: lockName},
				},
			},
		}, nil
	}
}
