package mock

import "github.com/FredyXue/go-utils/monitor"

var MockMonitor = NewMockMonitor()

func NewMockMonitor() monitor.Monitor {
	return &mockMonitor{}
}

type mockMonitor struct{}

func (m *mockMonitor) Start(group string)                                                   {}
func (m *mockMonitor) Stop()                                                                {}
func (m *mockMonitor) Register(method string, fn monitor.Callback, copt ...monitor.CallOpt) {}
func (m *mockMonitor) Deregister(method string)                                             {}

func (m *mockMonitor) Watch(method string, tag string, ctxData ...[]byte) (mctx monitor.MonitorContext, err error) {
	mctx = &mockMonitorContext{}
	return
}

func (m *mockMonitor) Unwatch(method string, tag string) error {
	return nil
}

func (m *mockMonitor) WatchList() (list []string, err error) {
	list = make([]string, 0)
	return
}

func (m *mockMonitor) IsMaster() bool {
	return true
}

type mockMonitorContext struct{}

func (c *mockMonitorContext) Get() ([]byte, error) {
	return make([]byte, 0), nil
}

func (c *mockMonitorContext) Set([]byte) error {
	return nil
}

func (c *mockMonitorContext) Check() (valid bool, err error) {
	return true, nil
}

func (c *mockMonitorContext) Close() error {
	return nil
}
