package fakes

import "sync"

type SimpleTesting struct {
	FatalfCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Format string
			Args   []interface {
			}
		}
		Stub func(string, ...interface {
		})
	}
	TempDirCall struct {
		mutex     sync.Mutex
		CallCount int
		Returns   struct {
			String string
		}
		Stub func() string
	}
}

func (f *SimpleTesting) Fatalf(param1 string, param2 ...interface {
}) {
	f.FatalfCall.mutex.Lock()
	defer f.FatalfCall.mutex.Unlock()
	f.FatalfCall.CallCount++
	f.FatalfCall.Receives.Format = param1
	f.FatalfCall.Receives.Args = param2
	if f.FatalfCall.Stub != nil {
		f.FatalfCall.Stub(param1, param2...)
	}
}
func (f *SimpleTesting) TempDir() string {
	f.TempDirCall.mutex.Lock()
	defer f.TempDirCall.mutex.Unlock()
	f.TempDirCall.CallCount++
	if f.TempDirCall.Stub != nil {
		return f.TempDirCall.Stub()
	}
	return f.TempDirCall.Returns.String
}
