package fakes

import "sync"

type RegistryBuildpackToLocal struct {
	ExtractCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Ref         string
			Destination string
		}
		Returns struct {
			String_1 string
			String_2 string
			Error    error
		}
		Stub func(string, string) (string, string, error)
	}
}

func (f *RegistryBuildpackToLocal) Extract(param1 string, param2 string) (string, string, error) {
	f.ExtractCall.mutex.Lock()
	defer f.ExtractCall.mutex.Unlock()
	f.ExtractCall.CallCount++
	f.ExtractCall.Receives.Ref = param1
	f.ExtractCall.Receives.Destination = param2
	if f.ExtractCall.Stub != nil {
		return f.ExtractCall.Stub(param1, param2)
	}
	return f.ExtractCall.Returns.String_1, f.ExtractCall.Returns.String_2, f.ExtractCall.Returns.Error
}
