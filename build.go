package noderunscript

import (
	"bytes"
	"os"
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

		buffer := bytes.NewBuffer(nil)
		mainExecutable := npmExec
		execution := pexec.Execution{
			Dir:    context.WorkingDir,
			Args:   []string{"run-script", ""},
			Stdout: buffer,
			Stderr: buffer,
		}

		packageManager := scriptManager.GetPackageManager(context.WorkingDir)

		if packageManager == "yarn" {
			mainExecutable = yarnExec
			execution.Args[0] = "run"
		}

		packageScripts, err := scriptManager.GetPackageScripts(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		envRunScripts, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.BuildResult{},
				packit.Fail.WithMessage("expected value from $BP_NODE_RUN_SCRIPTS to be set")
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		var scripts []string
		for _, envScriptName := range envScriptNames {
			if _, exists := packageScripts[envScriptName]; exists {
				scripts = append(scripts, envScriptName)
			}
		}

		logger.Process("Executing build process")
		logger.Subprocess("Executing scripts")

		duration, err := clock.Measure(func() error {
			for _, script := range scripts {
				logger.Action("Running '%s %s %s'", packageManager, execution.Args[0], script)

				execution.Args[1] = script
				err = mainExecutable.Execute(execution)

				logger.Detail("%s", buffer.String())
				buffer.Reset()

				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return packit.BuildResult{}, nil
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		return packit.BuildResult{}, nil
	}
}
