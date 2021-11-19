package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer"
)

type LocalFetcher struct {
	GetCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			LocalBuildpack freezer.LocalBuildpack
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(freezer.LocalBuildpack) (string, error)
	}
	WithPackagerCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Packager freezer.Packager
		}
		Returns struct {
			LocalFetcher freezer.LocalFetcher
		}
		Stub func(freezer.Packager) freezer.LocalFetcher
	}
}

func (f *LocalFetcher) Get(param1 freezer.LocalBuildpack) (string, error) {
	f.GetCall.mutex.Lock()
	defer f.GetCall.mutex.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.LocalBuildpack = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.String, f.GetCall.Returns.Error
}
func (f *LocalFetcher) WithPackager(param1 freezer.Packager) freezer.LocalFetcher {
	f.WithPackagerCall.mutex.Lock()
	defer f.WithPackagerCall.mutex.Unlock()
	f.WithPackagerCall.CallCount++
	f.WithPackagerCall.Receives.Packager = param1
	if f.WithPackagerCall.Stub != nil {
		return f.WithPackagerCall.Stub(param1)
	}
	return f.WithPackagerCall.Returns.LocalFetcher
}
