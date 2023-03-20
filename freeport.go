package portforwarder

import (
	"fmt"
	"net"
)

type netFreePortProvider struct {
	network string
	host    string
	port    uint
}

func newNetFreePortProvider(network string, host string, port uint) *netFreePortProvider {
	return &netFreePortProvider{network: network, host: host, port: port}
}

func (p *netFreePortProvider) getFreePort() (uint, error) {
	addr, err := net.ResolveTCPAddr(p.network, fmt.Sprintf("%s:%d", p.host, p.port))
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP(p.network, addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return uint(l.Addr().(*net.TCPAddr).Port), nil
}
