package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var settings struct {
	Buildpacks struct {
		NodeEngine struct {
			Online string
		}

		NpmInstall struct {
			Online string
		}

		Yarn struct {
			Online string
		}

		YarnInstall struct {
			Online string
		}

		NodeRunScript struct {
			Online string
		}
	}
	Buildpack struct {
		ID   string
		Name string
	}
	Config struct {
		NodeEngine  string `json:"node-engine"`
		NpmInstall  string `json:"npm-install"`
		Yarn        string `json:"yarn"`
		YarnInstall string `json:"yarn-install"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &settings.Buildpack)
	Expect(err).NotTo(HaveOccurred())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.NodeRunScript.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NpmInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.NpmInstall)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Yarn.Online, err = buildpackStore.Get.
		Execute(settings.Config.Yarn)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.YarnInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.YarnInstall)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("SimpleYarnApp", testSimpleYarnApp)
	suite("SimpleNPMApp", testSimpleNPMApp)
	suite("ProjectPathApp", testProjectPathApp)
	suite("VueNpmApp", testVueNpmApp)
	suite("VueYarnApp", testVueYarnApp)
	suite.Run(t)
}
