package transport

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/mf751/btenl/internal/shared/types"
)

// QuicConnection
// Implements types.Conn, it is peer-to-peer QUIC protocol connection
// returned from new connections can open and accept multiple streams
type QuicConnection struct {
	Conn *quic.Conn
	id   types.ConnID
}

// QuicConnector
// Implements the Connector interface which provides methods
// to open and receives connections. It uses the QUIC protocol
// with encrypted connection using TLS Certificate.
// the QUIC protocol provides multiple streams by default.
type QuicConnector struct {
	addr string
	cert *tls.Certificate
}

func NewQuicConnector(addr string, cert *tls.Certificate) *QuicConnector {
	return &QuicConnector{addr: addr, cert: cert}
}

func (ctr *QuicConnector) Listen(
	ctx context.Context,
	onConn func(types.Conn),
	onErr func(error),
) error {
	conf := &tls.Config{
		Certificates: []tls.Certificate{*ctr.cert},
		NextProtos:   []string{"btenl"},

		ClientAuth: tls.RequireAnyClientCert,

		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return errors.New("no peer Certificate")
			}

			if _, ok := state.PeerCertificates[0].PublicKey.(ed25519.PublicKey); !ok {
				return errors.New("peer Certificate Public Key must use ed25519")
			}

			return nil
		},
	}

	listener, err := quic.ListenAddr(ctr.addr, conf, nil)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			onErr(err)
			continue
		}

		go onConn(&QuicConnection{Conn: conn, id: types.ConnID(types.NextConnID())})
	}
}

func (ctr *QuicConnector) Connect(ctx context.Context, ip net.IP) (types.Conn, error) {
	conf := &tls.Config{
		Certificates:       []tls.Certificate{*ctr.cert},
		NextProtos:         []string{"btenl"},
		InsecureSkipVerify: true,

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

	if isSelfAddr(ip) {
		return nil, ErrOwnIPConnect
	}

	_, port, err := net.SplitHostPort(ctr.addr)
	if err != nil {
		return nil, err
	}

	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(dialCtx, net.JoinHostPort(ip.String(), port), conf, nil)
	if err != nil {
		return nil, err
	}

	return &QuicConnection{Conn: conn, id: types.ConnID(types.NextConnID())}, nil
}

func (c *QuicConnection) ID() types.ConnID {
	return c.id
}

func (c *QuicConnection) NodeID() types.NodeID {
	// Since Connection was successful that ensures PublicKey is ed25519
	cert := c.Conn.ConnectionState().TLS.PeerCertificates[0]
	return sha256.Sum256(cert.PublicKey.(ed25519.PublicKey))
}

func (c *QuicConnection) OpenStream(ctx context.Context) (types.Stream, error) {
	return c.Conn.OpenStreamSync(ctx)
}

func (c *QuicConnection) AcceptStream(ctx context.Context) (types.Stream, error) {
	return c.Conn.AcceptStream(ctx)
}

func (c *QuicConnection) Close() error {
	return c.Conn.CloseWithError(0, "")
}

func (c *QuicConnection) Ctx() context.Context {
	return c.Conn.Context()
}

func (c *QuicConnection) IP() net.IP {
	addr, ok := c.Conn.RemoteAddr().(*net.UDPAddr)
	if !ok {
		return nil
	}
	return addr.IP
}
