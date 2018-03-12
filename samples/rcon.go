// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	logrus "github.com/Sirupsen/logrus"
	"github.com/sostronk/go-steam"
)

var debug = flag.Bool("debug", false, "debug")

func main() {
	flag.Parse()

	if *debug {
		log := logrus.New()
		log.Level = logrus.DebugLevel
		steam.SetLog(log)
	}

	addr := os.Getenv("ADDR")
	pass := os.Getenv("RCON_PASSWORD")

	if addr == "" || pass == "" {
		fmt.Println("Please set ADDR & RCON_PASSWORD.")
		return
	}

	for {
		rcon, err := steam.Connect(addr, steam.WithRCONPassword(pass))
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
