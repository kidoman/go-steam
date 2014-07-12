package main

import (
	"flag"
	"fmt"

	"github.com/kidoman/go-steam"
)

func main() {
	flag.Parse()
	server := &steam.Server{Addr: "0.1.2.3:27015"}
	defer server.Close()

	authenticated, err := server.AuthenticateRcon("abc")
	if err != nil {
		panic(err)
	}
	fmt.Printf("authentication status %v\n", authenticated)

	comm := "status"
	resp, err := server.ExecRconCommand(comm)
	if err != nil {
		panic(err)
	}

	fmt.Printf("rcon command: %v  response: %v\n", comm, resp)
}
