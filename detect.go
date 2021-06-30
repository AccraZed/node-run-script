package noderunscript

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.DetectResult{},
				packit.Fail.WithMessage("environment variable $BP_NODE_RUN_SCRIPTS is not set")
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
