package fakes

import (
	"sync"

	"github.com/cloudfoundry/occam"
)

type DockerImageClient struct {
	InspectCall struct {
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

func (f *DockerImageClient) Inspect(param1 string) (occam.Image, error) {
	f.InspectCall.Lock()
	defer f.InspectCall.Unlock()
	f.InspectCall.CallCount++
	f.InspectCall.Receives.Ref = param1
	if f.InspectCall.Stub != nil {
		return f.InspectCall.Stub(param1)
	}
	return f.InspectCall.Returns.Image, f.InspectCall.Returns.Error
}
