package noderunscript_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	noderunscript "github.com/accrazed/node-run-script"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("BP_NODE_RUN_SCRIPTS", "build")

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`
			{
				"name": "mypackage",
				"scripts": {
				   "build": "mybuildcommand --args"
				}
			}`), 0644)).To(Succeed())

		detect = noderunscript.Detect()
	})

	it.After(func() {
		os.Unsetenv("BP_NODE_RUN_SCRIPTS")
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when the working-dir contains yarn.lock", func() {
		it("returns a plan that requires node and yarn", func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0644)).To(Succeed())

			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{Name: "node"}, {Name: "yarn"},
				},
			}))
		})
	})

	context("when the working-dir doesn't contain yarn.lock", func() {
		it("defaults to npm and returns a plan that requires node and npm", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{Name: "node"}, {Name: "npm"},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when the env var of \"$BP_NODE_RUN_SCRIPTS\" is not set", func() {
			it.Before(func() {
				os.Unsetenv("BP_NODE_RUN_SCRIPTS")
			})

			it("returns a failure", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("environment variable $BP_NODE_RUN_SCRIPTS is not set"))
			})
		})

		context("if package.json is absent", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())
			})

			it("returns a failure", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("file package.json does not exist"))
			})
		})

		context("if any of the scripts in \"$BP_NODE_RUN_SCRIPTS\" does not exist in package.json", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`
					{
						"name": "mypackage",
						"scripts": {
						"random-script": "mybuildcommand --args"
						}
					}`), 0644)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("one of the scripts in $BP_NODE_RUN_SCRIPTS does not exist in package.json"))
			})
		})
	})
}
