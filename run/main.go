package main

import (
	"os"

	noderunscript "github.com/accrazed/node-run-script"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	npmExec := pexec.NewExecutable("npm")
	yarnExec := pexec.NewExecutable("yarn")
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		noderunscript.Detect(),
		noderunscript.Build(npmExec, yarnExec, chronos.DefaultClock, logger),
	)
}
