package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer"
)

type RemoteFetcher struct {
	GetCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			RemoteBuildpack freezer.RemoteBuildpack
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(freezer.RemoteBuildpack) (string, error)
	}
}

func (f *RemoteFetcher) Get(param1 freezer.RemoteBuildpack) (string, error) {
	f.GetCall.Lock()
	defer f.GetCall.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.RemoteBuildpack = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.String, f.GetCall.Returns.Error
}
