package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/pexec"
)

type Executable struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Execution pexec.Execution
		}
		Returns struct {
			Error error
		}
		Stub func(pexec.Execution) error
	}
	NameCall struct {
		mutex     sync.Mutex
		CallCount int
		Returns   struct {
			String string
		}
		Stub func() string
	}
}

func (f *Executable) Execute(param1 pexec.Execution) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Execution = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Error
}
func (f *Executable) Name() string {
	f.NameCall.mutex.Lock()
	defer f.NameCall.mutex.Unlock()
	f.NameCall.CallCount++
	if f.NameCall.Stub != nil {
		return f.NameCall.Stub()
	}
	return f.NameCall.Returns.String
}
