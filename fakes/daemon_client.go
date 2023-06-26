package fakes

import (
	"context"
	"io"
	"sync"

	"github.com/docker/docker/api/types"
)

type DockerDaemonClient struct {
	ImageInspectWithRawCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context context.Context
			String  string
		}
		Returns struct {
			ImageInspect types.ImageInspect
			ByteSlice    []byte
			Error        error
		}
		Stub func(context.Context, string) (types.ImageInspect, []byte, error)
	}
	ImageLoadCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context context.Context
			Reader  io.Reader
			Bool    bool
		}
		Returns struct {
			ImageLoadResponse types.ImageLoadResponse
			Error             error
		}
		Stub func(context.Context, io.Reader, bool) (types.ImageLoadResponse, error)
	}
	ImageSaveCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context     context.Context
			StringSlice []string
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(context.Context, []string) (io.ReadCloser, error)
	}
	ImageTagCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context  context.Context
			String_1 string
			String_2 string
		}
		Returns struct {
			Error error
		}
		Stub func(context.Context, string, string) error
	}
	NegotiateAPIVersionCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Ctx context.Context
		}
		Stub func(context.Context)
	}
}

func (f *DockerDaemonClient) ImageInspectWithRaw(param1 context.Context, param2 string) (types.ImageInspect, []byte, error) {
	f.ImageInspectWithRawCall.mutex.Lock()
	defer f.ImageInspectWithRawCall.mutex.Unlock()
	f.ImageInspectWithRawCall.CallCount++
	f.ImageInspectWithRawCall.Receives.Context = param1
	f.ImageInspectWithRawCall.Receives.String = param2
	if f.ImageInspectWithRawCall.Stub != nil {
		return f.ImageInspectWithRawCall.Stub(param1, param2)
	}
	return f.ImageInspectWithRawCall.Returns.ImageInspect, f.ImageInspectWithRawCall.Returns.ByteSlice, f.ImageInspectWithRawCall.Returns.Error
}
func (f *DockerDaemonClient) ImageLoad(param1 context.Context, param2 io.Reader, param3 bool) (types.ImageLoadResponse, error) {
	f.ImageLoadCall.mutex.Lock()
	defer f.ImageLoadCall.mutex.Unlock()
	f.ImageLoadCall.CallCount++
	f.ImageLoadCall.Receives.Context = param1
	f.ImageLoadCall.Receives.Reader = param2
	f.ImageLoadCall.Receives.Bool = param3
	if f.ImageLoadCall.Stub != nil {
		return f.ImageLoadCall.Stub(param1, param2, param3)
	}
	return f.ImageLoadCall.Returns.ImageLoadResponse, f.ImageLoadCall.Returns.Error
}
func (f *DockerDaemonClient) ImageSave(param1 context.Context, param2 []string) (io.ReadCloser, error) {
	f.ImageSaveCall.mutex.Lock()
	defer f.ImageSaveCall.mutex.Unlock()
	f.ImageSaveCall.CallCount++
	f.ImageSaveCall.Receives.Context = param1
	f.ImageSaveCall.Receives.StringSlice = param2
	if f.ImageSaveCall.Stub != nil {
		return f.ImageSaveCall.Stub(param1, param2)
	}
	return f.ImageSaveCall.Returns.ReadCloser, f.ImageSaveCall.Returns.Error
}
func (f *DockerDaemonClient) ImageTag(param1 context.Context, param2 string, param3 string) error {
	f.ImageTagCall.mutex.Lock()
	defer f.ImageTagCall.mutex.Unlock()
	f.ImageTagCall.CallCount++
	f.ImageTagCall.Receives.Context = param1
	f.ImageTagCall.Receives.String_1 = param2
	f.ImageTagCall.Receives.String_2 = param3
	if f.ImageTagCall.Stub != nil {
		return f.ImageTagCall.Stub(param1, param2, param3)
	}
	return f.ImageTagCall.Returns.Error
}
func (f *DockerDaemonClient) NegotiateAPIVersion(param1 context.Context) {
	f.NegotiateAPIVersionCall.mutex.Lock()
	defer f.NegotiateAPIVersionCall.mutex.Unlock()
	f.NegotiateAPIVersionCall.CallCount++
	f.NegotiateAPIVersionCall.Receives.Ctx = param1
	if f.NegotiateAPIVersionCall.Stub != nil {
		f.NegotiateAPIVersionCall.Stub(param1)
	}
}
