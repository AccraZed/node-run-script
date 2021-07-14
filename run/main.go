package main

import (
	"os"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	npmExec := pexec.NewExecutable("npm")
	yarnExec := pexec.NewExecutable("yarn")
	scriptManager := noderunscript.NewScriptManager()
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		noderunscript.Detect(scriptManager),
		noderunscript.Build(npmExec, yarnExec, scriptManager, chronos.DefaultClock, logger),
	)
}
