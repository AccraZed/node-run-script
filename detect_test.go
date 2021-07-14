package noderunscript_test

import (
	"os"
	"path/filepath"
	"testing"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/paketo-buildpacks/node-run-script/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir    string
		detect        packit.DetectFunc
		scriptManager *fakes.PackageInterface
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Setenv("BP_NODE_RUN_SCRIPTS", "build")).To(Succeed())

		Expect(os.WriteFile(filepath.Join(workingDir, "package.json"), nil, 0644)).To(Succeed())

		scriptManager = &fakes.PackageInterface{}
		scriptManager.GetPackageScriptsCall.Returns.MapStringString = map[string]string{
			"build":       "mybuildcommand --args",
			"some-script": "somecommands --args",
		}

		detect = noderunscript.Detect(scriptManager)
	})

	it.After(func() {
		Expect(os.Unsetenv("BP_NODE_RUN_SCRIPTS")).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when the working-dir doesn't contain yarn.lock", func() {
		it.Before(func() {
			scriptManager.GetPackageManagerCall.Returns.String = "npm"
		})

		it("defaults to npm and returns a plan that requires node and npm", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "npm",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
				},
			}))
		})
	})

	context("when the working-dir contains yarn.lock", func() {
		it.Before(func() {
			scriptManager.GetPackageManagerCall.Returns.String = "yarn"
			Expect(os.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0644)).To(Succeed())
		})

		it("returns a plan that requires node and yarn", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "yarn",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
				},
			}))
		})
	})

	context("when env var $BP_NODE_RUN_SCRIPTS has spaces among commas", func() {
		it.Before(func() {
			scriptManager.GetPackageManagerCall.Returns.String = "npm"
			Expect(os.Setenv("BP_NODE_RUN_SCRIPTS", "build, some-script ")).To(Succeed())
		})

		it("trims the whitespace and successfully detects the scripts", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("when env var $BP_NODE_PROJECT_PATH is set", func() {
		it.Before(func() {
			Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())

			customPath, err := os.MkdirTemp(workingDir, "custom-project-path")
			Expect(err).NotTo(HaveOccurred())
			customPath = filepath.Base(customPath)
			Expect(os.Setenv("BP_NODE_PROJECT_PATH", customPath)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "yarn.lock"), nil, 0644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "package.json"), nil, 0644)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_NODE_PROJECT_PATH")).To(Succeed())
		})

		context("when the custom project path contains yarn.lock", func() {
			it.Before(func() {
				scriptManager.GetPackageManagerCall.Returns.String = "yarn"
			})

			it("returns a plan that requires node and yarn", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name:     "node",
							Metadata: noderunscript.BuildPlanMetadata{Build: true},
						},
						{
							Name:     "yarn",
							Metadata: noderunscript.BuildPlanMetadata{Build: true},
						},
					},
				}))
			})
		})
	})

	context("failure cases", func() {
		context("when the env var of \"$BP_NODE_RUN_SCRIPTS\" is not set", func() {
			it.Before(func() {
				Expect(os.Unsetenv("BP_NODE_RUN_SCRIPTS")).To(Succeed())
			})

			it("returns a failure", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("expected value from $BP_NODE_RUN_SCRIPTS to be set"))
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

				Expect(err).To(MatchError("expected file package.json to exist"))
			})
		})

		context("if any of the scripts in \"$BP_NODE_RUN_SCRIPTS\" does not exist in package.json", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_NODE_RUN_SCRIPTS", "build,script1,some-script,script2,script3")).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("expected script(s) script1, script2, script3 from $BP_NODE_RUN_SCRIPTS to exist in package.json"))
			})
		})

		context("if $BP_NODE_PROJECT_PATH leads to a directory that doesn't exist", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_NODE_PROJECT_PATH", "not_a_real_directory")).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("expected value from $BP_NODE_PROJECT_PATH [" + workingDir + "/not_a_real_directory] to be an existing directory"))
			})

		})
	})
}
