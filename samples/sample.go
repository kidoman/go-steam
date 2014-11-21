package main

import (
	"flag"
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
	flag.Parse()

	for _, address := range addresses {
		server := &steam.Server{Addr: address}
		ping, err := server.Ping()
		if err != nil {
			fmt.Printf("Could not ping %v\n", address)
			continue
		}
		info, err := server.Info()
		if err != nil {
			fmt.Printf("Could not get server info from %v\n", address)
			continue
		}
		playersInfo, err := server.PlayersInfo()
		if err != nil {
			fmt.Printf("Could not get players info from %v\n", address)
			continue
		}
		fmt.Printf("%v:\n%v with ping %v\n", address, info, ping)
		if len(playersInfo.Players) > 0 {
			fmt.Println("Player infos:")
			for _, player := range playersInfo.Players {
				fmt.Printf("%v %v\n", player.Name, player.Score)
			}
		}
		fmt.Println()
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
