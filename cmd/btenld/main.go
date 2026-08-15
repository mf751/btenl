package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mf751/btenl.git/internal/connections"
	"github.com/mf751/btenl.git/internal/controls"
	"github.com/mf751/btenl.git/internal/daemon"
	"github.com/mf751/btenl.git/internal/shared/encryption"
	"github.com/mf751/btenl.git/internal/shared/logger"
	"github.com/mf751/btenl.git/internal/transport"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logFile, err := os.OpenFile("/tmp/btenld.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0x644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger := logger.New(logFile)

	id, err := encryption.EnsureIdentity()
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("Identity ready: " + id.Fingerprint())

	cert, err := encryption.EnsureCertificate(id)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Infof("Certificate ready for: %v", id.Fingerprint())

	unixIPC := control.NewUnixIPCSource("/tmp/btenld.sock")

	quicConnector := transport.NewQuicConnector("0.0.0.0:9001", cert)

	connectionMux := connections.NewConnectionMux(quicConnector)

	d := daemon.New(ctx, connectionMux, logger)

	d.ForwardControlSource(unixIPC)
	d.ForwardConnectionSource(connectionMux)

	d.Run()
}
