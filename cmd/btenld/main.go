package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mf751/btenl.git/internal/controls"
	"github.com/mf751/btenl.git/internal/daemon"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	unixIPC := control.NewUnixIPCSource("/tmp/btenld.sock")

	d := daemon.New(ctx)

	d.ForwardControlSource(unixIPC)

	d.Run()
}
