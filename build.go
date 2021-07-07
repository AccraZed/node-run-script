package noderunscript

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) error
}

func Build(npmExec Executable, yarnExec Executable, scriptManager PackageInterface, clock chronos.Clock, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		projectDir := context.WorkingDir
		bpNodeProjectPath, exists := os.LookupEnv("BP_NODE_PROJECT_PATH")
		if exists {
			projectDir = filepath.Join(context.WorkingDir, bpNodeProjectPath)
		}

		buffer := bytes.NewBuffer(nil)
		mainExecutable := npmExec
		execution := pexec.Execution{
			Dir:    projectDir,
			Args:   []string{"run-script", ""},
			Stdout: buffer,
			Stderr: buffer,
		}

		packageManager := scriptManager.GetPackageManager(projectDir)

		if packageManager == "yarn" {
			mainExecutable = yarnExec
			execution.Args[0] = "run"
		}

		scripts := strings.Split(os.Getenv("BP_NODE_RUN_SCRIPTS"), ",")

		logger.Process("Executing build process")
		logger.Subprocess("Executing scripts")

		duration, err := clock.Measure(func() error {
			for _, script := range scripts {
				logger.Action("Running '%s %s %s'", packageManager, execution.Args[0], script)

				execution.Args[1] = script
				err := mainExecutable.Execute(execution)

				logger.Detail("%s", buffer)
				buffer.Reset()

				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		return packit.BuildResult{}, nil
	}
}
