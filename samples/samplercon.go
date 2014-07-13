package main

import (
	"flag"
	"fmt"

	"github.com/kidoman/go-steam"
)

func main() {
	flag.Parse()
	server := &steam.Server{Addr: "192.168.59.103:27015"}
	defer server.Close()

	authenticated, err := server.AuthenticateRcon("saPa3igYuForTs3Xbbmo")
	if err != nil {
		panic(err)
	}
	fmt.Printf("authentication status %v\n", authenticated)

	comm := "quit"
	resp, err := server.ExecRconCommand(comm)
	if err != nil {
		panic(err)
	}

	fmt.Printf("rcon command: %v  response: %v\n", comm, resp)
}
