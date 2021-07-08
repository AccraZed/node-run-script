package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testProjectPathApp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack occam.Pack
	)

	it.Before(func() {
		pack = occam.NewPack()
	})

	context("when building a simple yarn app inside a nested directory", func() {
		var (
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it("builds an OCI image for the app", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "project_path_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			_, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.NodeRunScript.Online,
				).
				WithEnv(map[string]string{
					"BP_NODE_RUN_SCRIPTS":  "test_script_1,test_script_2",
					"BP_NODE_PROJECT_PATH": "nested_yarn_app"}).
				WithPullPolicy("never").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Executing build process",
				"    Executing scripts",
				"      Running 'yarn run test_script_1'",
				"      Running 'yarn run test_script_2'",
			))
			Expect(logs).To(ContainLines(MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`)))
		})
	})
}
