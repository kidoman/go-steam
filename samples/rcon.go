package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/kidoman/go-steam"
)

func main() {
	debug := flag.Bool("debug", false, "debug")
	flag.Parse()
	if *debug {
		steam.SetLog(log.New())
	}
	addr := os.Getenv("ADDR")
	pass := os.Getenv("RCON_PASSWORD")
	if addr == "" || pass == "" {
		fmt.Println("Please set ADDR & RCON_PASSWORD.")
		return
	}
	for {
		o := &steam.ConnectOptions{RCONPassword: pass}
		rcon, err := steam.Connect(addr, o)
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer rcon.Close()
		for {
			resp, err := rcon.Send("status")
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(resp)
			time.Sleep(5 * time.Second)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
