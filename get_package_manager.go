package noderunscript

import (
	"os"
	"path/filepath"
)

func getPackageManager(workingDir string) string {
	_, err := os.Stat(filepath.Join(workingDir, "yarn.lock"))
	if err == nil {
		return "yarn"
	}
	return "npm"
}
