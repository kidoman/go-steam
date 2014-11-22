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

	for _, addr := range addresses {
		server := &steam.Server{Addr: addr, RCONPassword: "abc"}
		ping, err := server.Ping()
		if err != nil {
			fmt.Printf("steam: could not ping %v: %v", addr, err)
			continue
		}
		info, err := server.Info()
		if err != nil {
			fmt.Printf("steam: could not get server info from %v: %v", addr, err)
			continue
		}
		playersInfo, err := server.PlayersInfo()
		if err != nil {
			fmt.Printf("steam: could not get players info from %v: %v", addr, err)
			continue
		}
		fmt.Printf("%v:\n%v with ping %v\n", addr, info, ping)
		if len(playersInfo.Players) > 0 {
			fmt.Println("steam: player infos:")
			for _, player := range playersInfo.Players {
				fmt.Printf("steam: %v %v\n", player.Name, player.Score)
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
