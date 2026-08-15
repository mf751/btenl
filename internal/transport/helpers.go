package transport

import (
	"errors"
	"net"
)

var ErrOwnIPConnect = errors.New("can't connect to own ip")

// Checks if ip points back to own machine
func isSelfAddr(ip net.IP) bool {
	// 127.x.x.x stuff
	if ip.IsLoopback() {
		return true
	}

	// other ip's that point to own machine
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ip.Equal(ipNet.IP) {
			return true
		}
	}

	return false
}
