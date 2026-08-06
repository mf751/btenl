package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	control "github.com/mf751/btenl.git/internal/control_source"
	"github.com/mf751/btenl.git/internal/daemon"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	unixIPC := control.NewUnixIPCSource("/tmp/btenld.sock")

	controlMux := control.New(unixIPC)

	d := daemon.New(ctx, controlMux)

	d.Run()
}
