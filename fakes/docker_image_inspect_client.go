package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/occam"
)

type DockerImageInspectClient struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Ref string
		}
		Returns struct {
			Image occam.Image
			Error error
		}
		Stub func(string) (occam.Image, error)
	}
}

func (f *DockerImageInspectClient) Execute(param1 string) (occam.Image, error) {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Ref = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Image, f.ExecuteCall.Returns.Error
}
