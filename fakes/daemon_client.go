package fakes

import (
	"context"
	"io"
	"sync"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerDaemonClient struct {
	ImageHistoryCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context                 context.Context
			String                  string
			ImageHistoryOptionSlice []client.ImageHistoryOption
		}
		Returns struct {
			HistoryResponseItemSlice []image.HistoryResponseItem
			Error                    error
		}
		Stub func(context.Context, string, ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error)
	}
	ImageInspectWithRawCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context context.Context
			String  string
		}
		Returns struct {
			InspectResponse image.InspectResponse
			ByteSlice       []byte
			Error           error
		}
		Stub func(context.Context, string) (image.InspectResponse, []byte, error)
	}
	ImageLoadCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context              context.Context
			Reader               io.Reader
			ImageLoadOptionSlice []client.ImageLoadOption
		}
		Returns struct {
			LoadResponse image.LoadResponse
			Error        error
		}
		Stub func(context.Context, io.Reader, ...client.ImageLoadOption) (image.LoadResponse, error)
	}
	ImageSaveCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context              context.Context
			StringSlice          []string
			ImageSaveOptionSlice []client.ImageSaveOption
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(context.Context, []string, ...client.ImageSaveOption) (io.ReadCloser, error)
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

func (f *DockerDaemonClient) ImageHistory(param1 context.Context, param2 string, param3 ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error) {
	f.ImageHistoryCall.mutex.Lock()
	defer f.ImageHistoryCall.mutex.Unlock()
	f.ImageHistoryCall.CallCount++
	f.ImageHistoryCall.Receives.Context = param1
	f.ImageHistoryCall.Receives.String = param2
	f.ImageHistoryCall.Receives.ImageHistoryOptionSlice = param3
	if f.ImageHistoryCall.Stub != nil {
		return f.ImageHistoryCall.Stub(param1, param2, param3...)
	}
	return f.ImageHistoryCall.Returns.HistoryResponseItemSlice, f.ImageHistoryCall.Returns.Error
}
func (f *DockerDaemonClient) ImageInspectWithRaw(param1 context.Context, param2 string) (image.InspectResponse, []byte, error) {
	f.ImageInspectWithRawCall.mutex.Lock()
	defer f.ImageInspectWithRawCall.mutex.Unlock()
	f.ImageInspectWithRawCall.CallCount++
	f.ImageInspectWithRawCall.Receives.Context = param1
	f.ImageInspectWithRawCall.Receives.String = param2
	if f.ImageInspectWithRawCall.Stub != nil {
		return f.ImageInspectWithRawCall.Stub(param1, param2)
	}
	return f.ImageInspectWithRawCall.Returns.InspectResponse, f.ImageInspectWithRawCall.Returns.ByteSlice, f.ImageInspectWithRawCall.Returns.Error
}
func (f *DockerDaemonClient) ImageLoad(param1 context.Context, param2 io.Reader, param3 ...client.ImageLoadOption) (image.LoadResponse, error) {
	f.ImageLoadCall.mutex.Lock()
	defer f.ImageLoadCall.mutex.Unlock()
	f.ImageLoadCall.CallCount++
	f.ImageLoadCall.Receives.Context = param1
	f.ImageLoadCall.Receives.Reader = param2
	f.ImageLoadCall.Receives.ImageLoadOptionSlice = param3
	if f.ImageLoadCall.Stub != nil {
		return f.ImageLoadCall.Stub(param1, param2, param3...)
	}
	return f.ImageLoadCall.Returns.LoadResponse, f.ImageLoadCall.Returns.Error
}
func (f *DockerDaemonClient) ImageSave(param1 context.Context, param2 []string, param3 ...client.ImageSaveOption) (io.ReadCloser, error) {
	f.ImageSaveCall.mutex.Lock()
	defer f.ImageSaveCall.mutex.Unlock()
	f.ImageSaveCall.CallCount++
	f.ImageSaveCall.Receives.Context = param1
	f.ImageSaveCall.Receives.StringSlice = param2
	f.ImageSaveCall.Receives.ImageSaveOptionSlice = param3
	if f.ImageSaveCall.Stub != nil {
		return f.ImageSaveCall.Stub(param1, param2, param3...)
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
