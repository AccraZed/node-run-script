package noderunscript

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

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
