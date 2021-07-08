package noderunscript_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	noderunscript "github.com/accrazed/node-run-script"
	"github.com/accrazed/node-run-script/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir   string
		workingDir  string
		projectPath string
		cnbDir      string

		build packit.BuildFunc

		timestamp     time.Time
		logger        scribe.Logger
		loggerBuffer  *bytes.Buffer
		npmExec       *fakes.Executable
		yarnExec      *fakes.Executable
		scriptManager *fakes.PackageInterface
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("BP_NODE_RUN_SCRIPTS", "build,some-script")
		Expect(ioutil.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`
			{
				"name": "mypackage",
				"scripts": {
				   "build": "echo \"script build running!\"",
				   "some-script": "echo \"script some-script running!\""
				}
			}`), 0644)).To(Succeed())

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		npmExec = &fakes.Executable{}
		yarnExec = &fakes.Executable{}
		scriptManager = &fakes.PackageInterface{}
		loggerBuffer = bytes.NewBuffer(nil)
		logger = scribe.NewLogger(loggerBuffer)

		build = noderunscript.Build(npmExec, yarnExec, scriptManager, clock, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is no yarn.lock", func() {
		it("runs npm commands", func() {
			npmExec.ExecuteCall.Returns.Error = nil

			scriptManager.GetPackageManagerCall.Returns.String = "npm"
			scriptManager.GetPackageScriptsCall.Returns.MapStringString = map[string]string{
				"build":       "echo \"script build running!\"",
				"some-script": "echo \"script some-script running!\"",
			}

			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(npmExec.ExecuteCall.CallCount).To(Equal(2))
			Expect(npmExec.ExecuteCall.Receives.Execution.Args).To(
				Equal([]string{"run-script", "some-script"}))
			Expect(npmExec.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		})
	})

	context("when there is a yarn.lock", func() {
		it("runs yarn commands", func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0644)).To(Succeed())
			yarnExec.ExecuteCall.Returns.Error = nil

			scriptManager.GetPackageManagerCall.Returns.String = "yarn"
			scriptManager.GetPackageScriptsCall.Returns.MapStringString = map[string]string{
				"build":       "echo \"script build running!\"",
				"some-script": "echo \"script some-script running!\"",
			}

			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(yarnExec.ExecuteCall.CallCount).To(Equal(2))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Args).To(
				Equal([]string{"run", "some-script"}))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		})
	})

	context("when there is a custom project path set", func() {
		it.Before(func() {
			var err error
			projectPath, err = os.MkdirTemp(workingDir, "custom-project-path")
			Expect(err).NotTo(HaveOccurred())

			customPath := filepath.Base(projectPath)
			os.Setenv("BP_NODE_PROJECT_PATH", customPath)

			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "yarn.lock"), nil, 0644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "package.json"), []byte(`
			{
				"name": "mypackage",
				"scripts": {
					"build": "mybuildcommand --args"
				}
				}`), 0644)).To(Succeed())

			os.Setenv("BP_NODE_RUN_SCRIPTS", "build")
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_PROJECT_PATH")
		})

		it("works and runs the correct commands", func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0644)).To(Succeed())
			yarnExec.ExecuteCall.Returns.Error = nil

			scriptManager.GetPackageManagerCall.Returns.String = "yarn"
			scriptManager.GetPackageScriptsCall.Returns.MapStringString = map[string]string{
				"build": "mybuildcommand --args",
			}

			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(yarnExec.ExecuteCall.CallCount).To(Equal(1))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Args).To(
				Equal([]string{"run", "build"}))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Dir).To(Equal(projectPath))
		})
	})

	context("failure cases", func() {
		context("when the script getting run has an error", func() {
			it("returns an error", func() {
				scriptManager.GetPackageManagerCall.Returns.String = "npm"
				npmExec.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, "some stdout output")
					fmt.Fprintln(execution.Stderr, "some stderr output")

					return fmt.Errorf("some execute error")
				}

				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})

				Expect(loggerBuffer.String()).To(ContainSubstring("some stdout output"))
				Expect(loggerBuffer.String()).To(ContainSubstring("some stderr output"))
				Expect(err).To(MatchError("some execute error"))
			})
		})
	})
}
