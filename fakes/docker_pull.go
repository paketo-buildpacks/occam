package fakes

import "sync"

type BPStoreDockerPull struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			String string
		}
		Returns struct {
			Error error
		}
		Stub func(string) error
	}
}

func (f *BPStoreDockerPull) Execute(param1 string) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.String = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Error
}
