package internal

import (
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"testing"
)

func TestNopResolver(t *testing.T) {
	// make sure ResolveNow & Close don't panic
	var r nopResolver
	r.ResolveNow(resolver.ResolveNowOptions{})
	r.Close()
}

type mockedClientConn struct {
	state resolver.State
	err   error
}

func (m *mockedClientConn) UpdateState(state resolver.State) error {
	m.state = state
	return m.err
}

func (m *mockedClientConn) ReportError(err error) {
}

func (m *mockedClientConn) NewAddress(addresses []resolver.Address) {
}

func (m *mockedClientConn) NewServiceConfig(serviceConfig string) {
}

func (m *mockedClientConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return nil
}
