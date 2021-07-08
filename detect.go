package noderunscript

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

type BuildPlanMetadata struct {
	Build bool `toml:"build"`
}

func Detect(scriptManager PackageInterface) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		envRunScripts, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.DetectResult{},
				packit.Fail.WithMessage("expected value from $BP_NODE_RUN_SCRIPTS to be set")
		}

		projectDir := context.WorkingDir
		envProjectPath, exists := os.LookupEnv("BP_NODE_PROJECT_PATH")
		if exists {
			projectDir = filepath.Join(context.WorkingDir, envProjectPath)
		}
		projectDir = filepath.Clean(projectDir)

		_, err := os.Stat(projectDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return packit.DetectResult{},
					fmt.Errorf("expected value from $BP_NODE_PROJECT_PATH [%s] to be an existing directory",
						projectDir)
			}
			return packit.DetectResult{}, err
		}

		_, err = os.Stat(filepath.Join(projectDir, "package.json"))
		if err != nil {
			return packit.DetectResult{},
				packit.Fail.WithMessage("expected file package.json to exist")
		}

		packageScripts, err := scriptManager.GetPackageScripts(projectDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		envScripts := strings.Split(envRunScripts, ",")

		for _, envScript := range envScripts {
			if _, exists := packageScripts[envScript]; !exists {
				return packit.DetectResult{},
					fmt.Errorf("expected a script from $BP_NODE_RUN_SCRIPTS to exist in package.json")
			}
		}

		packageManager := scriptManager.GetPackageManager(projectDir)

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: BuildPlanMetadata{Build: true},
					},
					{
						Name:     packageManager,
						Metadata: BuildPlanMetadata{Build: true},
					},
				},
			},
		}, nil
	}
}
