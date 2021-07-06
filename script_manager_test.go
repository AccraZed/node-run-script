package noderunscript_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	noderunscript "github.com/accrazed/node-run-script"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testScriptManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir    string
		scriptManager *noderunscript.ScriptManager
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		scriptManager = noderunscript.CreateScriptManager()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("GetPackageScripts", func() {

		it("succeeds", func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`
			{
				"name": "mypackage",
				"scripts": {
				   "build": "mybuildcommand --args",
				   "some-script": "somecommands --args"
				}
			}`), 0644)).To(Succeed())

			result, err := scriptManager.GetPackageScripts(workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["build"]).To(Equal("mybuildcommand --args"))
			Expect(result["some-script"]).To(Equal("somecommands --args"))
		})

		context("failure cases", func() {
			context("when package.json is incorrectly formatted", func() {
				it("fails", func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`
					{
						"name": "mypackage"
						"build": "mybuildcommand --args"
						"some-script": "somecommands --args"
						}`), 0644)).To(Succeed())

					_, err := scriptManager.GetPackageScripts(workingDir)
					Expect(err).To(HaveOccurred())

				})
			})
			context("when package.json does not exist", func() {
				it("fails", func() {
					_, err := scriptManager.GetPackageScripts(workingDir)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	context("GetPackageManager", func() {
		context("when yarn.lock exists", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0644)).To(Succeed())
			})
			it("returns 'yarn'", func() {
				result := scriptManager.GetPackageManager(workingDir)
				Expect(result).To(Equal("yarn"))
			})
		})

		context("when yarn.lock doesn't exist", func() {
			it("returns 'npm'", func() {
				result := scriptManager.GetPackageManager(workingDir)
				Expect(result).To(Equal("npm"))
			})
		})
	})
}
