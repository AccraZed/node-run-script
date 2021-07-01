package main

import (
	noderunscript "github.com/accrazed/node-run-script"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
)

func main() {

	packit.Run(
		noderunscript.Detect(),
		noderunscript.Build(
			pexec.NewExecutable("npm"),
			pexec.NewExecutable("yarn"),
		),
	)
}
