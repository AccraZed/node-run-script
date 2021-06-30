package noderunscript_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNodeRunScript(t *testing.T) {
	suite := spec.New("node-run-script", spec.Report(report.Terminal{}), spec.Sequential())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite.Run(t)
}
