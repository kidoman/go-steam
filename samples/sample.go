package main

import (
	"fmt"

	"github.com/kidoman/go-steam"
)

var addresses = []string{
	"116.251.223.240:27015",
	"116.251.223.240:10003",
	"119.81.28.246:27015",
	"116.251.216.147:27015",
	"119.81.28.243:27015",
}

func main() {
	for _, address := range addresses {
		server := &steam.Server{Addr: address}
		ping, err := server.Ping()
		must(err)
		info, err := server.Info()
		must(err)
		fmt.Printf("%v: %v with ping %v\n", address, info, ping)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
