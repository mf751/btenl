package transport

import (
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/quic-go/quic-go"
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

func TLSConf(cert *tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		NextProtos:         []string{"btenl"},
		InsecureSkipVerify: true,
		ClientAuth:         tls.RequireAnyClientCert,

		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return errors.New("no peer Certificate")
			}

			if _, ok := state.PeerCertificates[0].PublicKey.(ed25519.PublicKey); !ok {
				return errors.New("peer Certificate must use ed25519")
			}

			return nil
		},
	}
}

// Keep the connection alive send packets every 10 seconds
func QuicConf() *quic.Config {
	return &quic.Config{MaxIdleTimeout: 5 * time.Minute, KeepAlivePeriod: 10 * time.Second}
}
