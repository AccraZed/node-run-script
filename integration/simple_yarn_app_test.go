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

func testSimpleYarnApp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack occam.Pack
	)

	it.Before(func() {
		pack = occam.NewPack()
	})

	context("when building a simple yarn app", func() {
		var (
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it("builds an OCI image for a simple yarn app", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "simple_yarn_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			_, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.NodeRunScript.Online,
				).
				WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "test_script_1,test_script_2"}).
				WithPullPolicy("never").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Executing build process",
				"    Executing scripts",
			))
			Expect(logs).To(ContainLines("        some commands"))
			Expect(logs).To(ContainLines("        some other commands"))
			Expect(logs).To(ContainLines(MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`)))
		})
	})
}
