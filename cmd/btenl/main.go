package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type Response struct {
	Status  uint8  `json:"status"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: btenl <command> [args...]")
		os.Exit(1)
		return
	}

	command := os.Args[1]

	switch command {
	case "help":
		fmt.Println("")
		return
	case "version":
		fmt.Println("1.0")
		return
	}

	conn, err := net.Dial("unix", "/tmp/btenld.sock")
	if err != nil {
		fmt.Println(`
daemon is not running
run "btenl start" to run daemon
			`)
		os.Exit(1)
	}
	defer conn.Close()

	req := Request{
		Command: os.Args[1],
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
