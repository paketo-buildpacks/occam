package fakes

import (
	"context"
	"io"
	"sync"

	"github.com/moby/moby/client"
)

type DockerDaemonClient struct {
	PingCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context     context.Context
			PingOptions client.PingOptions
		}
		Returns struct {
			PingResult client.PingResult
			Error      error
		}
		Stub func(context.Context, client.PingOptions) (client.PingResult, error)
	}
	ImageHistoryCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context                 context.Context
			String                  string
			ImageHistoryOptionSlice []client.ImageHistoryOption
		}
		Returns struct {
			ImageHistoryResult client.ImageHistoryResult
			Error              error
		}
		Stub func(context.Context, string, ...client.ImageHistoryOption) (client.ImageHistoryResult, error)
	}
	ImageInspectCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context                context.Context
			String                 string
			ImageInspectOptionSlice []client.ImageInspectOption
		}
		Returns struct {
			ImageInspectResult client.ImageInspectResult
			Error              error
		}
		Stub func(context.Context, string, ...client.ImageInspectOption) (client.ImageInspectResult, error)
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
			ImageLoadResult client.ImageLoadResult
			Error           error
		}
		Stub func(context.Context, io.Reader, ...client.ImageLoadOption) (client.ImageLoadResult, error)
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
			ImageSaveResult client.ImageSaveResult
			Error           error
		}
		Stub func(context.Context, []string, ...client.ImageSaveOption) (client.ImageSaveResult, error)
	}
	ImageTagCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context         context.Context
			ImageTagOptions client.ImageTagOptions
		}
		Returns struct {
			ImageTagResult client.ImageTagResult
			Error          error
		}
		Stub func(context.Context, client.ImageTagOptions) (client.ImageTagResult, error)
	}
}

func (f *DockerDaemonClient) Ping(param1 context.Context, param2 client.PingOptions) (client.PingResult, error) {
	f.PingCall.mutex.Lock()
	defer f.PingCall.mutex.Unlock()
	f.PingCall.CallCount++
	f.PingCall.Receives.Context = param1
	f.PingCall.Receives.PingOptions = param2
	if f.PingCall.Stub != nil {
		return f.PingCall.Stub(param1, param2)
	}
	return f.PingCall.Returns.PingResult, f.PingCall.Returns.Error
}

func (f *DockerDaemonClient) ImageHistory(param1 context.Context, param2 string, param3 ...client.ImageHistoryOption) (client.ImageHistoryResult, error) {
	f.ImageHistoryCall.mutex.Lock()
	defer f.ImageHistoryCall.mutex.Unlock()
	f.ImageHistoryCall.CallCount++
	f.ImageHistoryCall.Receives.Context = param1
	f.ImageHistoryCall.Receives.String = param2
	f.ImageHistoryCall.Receives.ImageHistoryOptionSlice = param3
	if f.ImageHistoryCall.Stub != nil {
		return f.ImageHistoryCall.Stub(param1, param2, param3...)
	}
	return f.ImageHistoryCall.Returns.ImageHistoryResult, f.ImageHistoryCall.Returns.Error
}

func (f *DockerDaemonClient) ImageInspect(param1 context.Context, param2 string, param3 ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	f.ImageInspectCall.mutex.Lock()
	defer f.ImageInspectCall.mutex.Unlock()
	f.ImageInspectCall.CallCount++
	f.ImageInspectCall.Receives.Context = param1
	f.ImageInspectCall.Receives.String = param2
	f.ImageInspectCall.Receives.ImageInspectOptionSlice = param3
	if f.ImageInspectCall.Stub != nil {
		return f.ImageInspectCall.Stub(param1, param2, param3...)
	}
	return f.ImageInspectCall.Returns.ImageInspectResult, f.ImageInspectCall.Returns.Error
}

func (f *DockerDaemonClient) ImageLoad(param1 context.Context, param2 io.Reader, param3 ...client.ImageLoadOption) (client.ImageLoadResult, error) {
	f.ImageLoadCall.mutex.Lock()
	defer f.ImageLoadCall.mutex.Unlock()
	f.ImageLoadCall.CallCount++
	f.ImageLoadCall.Receives.Context = param1
	f.ImageLoadCall.Receives.Reader = param2
	f.ImageLoadCall.Receives.ImageLoadOptionSlice = param3
	if f.ImageLoadCall.Stub != nil {
		return f.ImageLoadCall.Stub(param1, param2, param3...)
	}
	return f.ImageLoadCall.Returns.ImageLoadResult, f.ImageLoadCall.Returns.Error
}

func (f *DockerDaemonClient) ImageSave(param1 context.Context, param2 []string, param3 ...client.ImageSaveOption) (client.ImageSaveResult, error) {
	f.ImageSaveCall.mutex.Lock()
	defer f.ImageSaveCall.mutex.Unlock()
	f.ImageSaveCall.CallCount++
	f.ImageSaveCall.Receives.Context = param1
	f.ImageSaveCall.Receives.StringSlice = param2
	f.ImageSaveCall.Receives.ImageSaveOptionSlice = param3
	if f.ImageSaveCall.Stub != nil {
		return f.ImageSaveCall.Stub(param1, param2, param3...)
	}
	return f.ImageSaveCall.Returns.ImageSaveResult, f.ImageSaveCall.Returns.Error
}

func (f *DockerDaemonClient) ImageTag(param1 context.Context, param2 client.ImageTagOptions) (client.ImageTagResult, error) {
	f.ImageTagCall.mutex.Lock()
	defer f.ImageTagCall.mutex.Unlock()
	f.ImageTagCall.CallCount++
	f.ImageTagCall.Receives.Context = param1
	f.ImageTagCall.Receives.ImageTagOptions = param2
	if f.ImageTagCall.Stub != nil {
		return f.ImageTagCall.Stub(param1, param2)
	}
	return f.ImageTagCall.Returns.ImageTagResult, f.ImageTagCall.Returns.Error
}
