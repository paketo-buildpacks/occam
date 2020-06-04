package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer"
)

type LocalFetcher struct {
	GetCall struct {
		sync.Mutex
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
}

func (f *LocalFetcher) Get(param1 freezer.LocalBuildpack) (string, error) {
	f.GetCall.Lock()
	defer f.GetCall.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.LocalBuildpack = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.String, f.GetCall.Returns.Error
}
