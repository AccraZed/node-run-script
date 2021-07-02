package fakes

import "sync"

type PackageInterface struct {
	GetPackageManagerCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			String string
		}
		Stub func(string) string
	}
	GetPackageScriptsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			MapStringString map[string]string
			Error           error
		}
		Stub func(string) (map[string]string, error)
	}
}

func (f *PackageInterface) GetPackageManager(param1 string) string {
	f.GetPackageManagerCall.Lock()
	defer f.GetPackageManagerCall.Unlock()
	f.GetPackageManagerCall.CallCount++
	f.GetPackageManagerCall.Receives.WorkingDir = param1
	if f.GetPackageManagerCall.Stub != nil {
		return f.GetPackageManagerCall.Stub(param1)
	}
	return f.GetPackageManagerCall.Returns.String
}
func (f *PackageInterface) GetPackageScripts(param1 string) (map[string]string, error) {
	f.GetPackageScriptsCall.Lock()
	defer f.GetPackageScriptsCall.Unlock()
	f.GetPackageScriptsCall.CallCount++
	f.GetPackageScriptsCall.Receives.WorkingDir = param1
	if f.GetPackageScriptsCall.Stub != nil {
		return f.GetPackageScriptsCall.Stub(param1)
	}
	return f.GetPackageScriptsCall.Returns.MapStringString, f.GetPackageScriptsCall.Returns.Error
}
