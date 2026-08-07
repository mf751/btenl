package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	socketPath = "/tmp/btenld.sock"
	logPath    = "/tmp/btenld.log"
)

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type Response struct {
	Status  uint8  `json:"status"`
	Message string `json:"message"`
}

func daemonRunning() bool {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func findDaemonBinary() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		bin := filepath.Join(filepath.Dir(exe), "btenld")
		if _, err := os.Stat(bin); err == nil {
			return bin, nil
		}
	}
	return exec.LookPath("btenld")
}

func start() {
	if daemonRunning() {
		fmt.Println("deamon is already running")
		return
	}

	bin, err := findDaemonBinary()
	if err != nil {
		fmt.Println("daemon binary not found:", err)
		os.Exit(1)
	}

	cmd := exec.Command(bin)
	// NOTE: new session so terminal signals don't kill the daemon
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("failed to open log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		fmt.Println("failed to start daemon:", err)
		os.Exit(1)
	}

	// NOTE: wait for the daemon to create its socket
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if daemonRunning() {
			fmt.Printf("daemon started successfuly (pid %d)\n", cmd.Process.Pid)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	cmd.Process.Kill()
	fmt.Println("daemon failed to start; check", logPath)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: btenl <command> [args...]")
		os.Exit(1)
		return
	}

	command := os.Args[1]

	switch command {
	case "start":
		start()
		return
	case "help":
		fmt.Println("")
		return
	case "version":
		fmt.Println("1.0")
		return
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println(`
daemon is not running
run "btenl start" to run daemon
			`)
		os.Exit(1)
	}
	defer conn.Close()

	req := Request{
		Command: command,
		Args:    os.Args[2:],
	}

	err = json.NewEncoder(conn).Encode(req)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(conn)

	for {
		var res Response

		if err = dec.Decode(&res); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println(err)
			return
		}

		fmt.Println(res.Message)

		if res.Status != 0 {
			return
		}
	}
}
