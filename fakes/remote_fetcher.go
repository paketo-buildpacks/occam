package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer"
)

type RemoteFetcher struct {
	GetCall struct {
		mutex     sync.Mutex
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
	WithPackagerCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Packager freezer.Packager
		}
		Returns struct {
			RemoteFetcher freezer.RemoteFetcher
		}
		Stub func(freezer.Packager) freezer.RemoteFetcher
	}
}

func (f *RemoteFetcher) Get(param1 freezer.RemoteBuildpack) (string, error) {
	f.GetCall.mutex.Lock()
	defer f.GetCall.mutex.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.RemoteBuildpack = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.String, f.GetCall.Returns.Error
}
func (f *RemoteFetcher) WithPackager(param1 freezer.Packager) freezer.RemoteFetcher {
	f.WithPackagerCall.mutex.Lock()
	defer f.WithPackagerCall.mutex.Unlock()
	f.WithPackagerCall.CallCount++
	f.WithPackagerCall.Receives.Packager = param1
	if f.WithPackagerCall.Stub != nil {
		return f.WithPackagerCall.Stub(param1)
	}
	return f.WithPackagerCall.Returns.RemoteFetcher
}
