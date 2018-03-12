package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sostronk/go-steam"
	"github.com/urfave/cli"
)

func init() {
	registerCommand(&pingCommand)
}

var pingCommand = cli.Command{
	Name:  "ping",
	Usage: "ping a server",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "v",
			Usage: "print server info",
		},
		cli.BoolFlag{
			Name:  "vv",
			Usage: "print more server info",
		},
		cli.BoolFlag{
			Name:  "forever, f",
			Usage: "ping forever",
		},
		cli.DurationFlag{
			Name:  "delay, d",
			Usage: "time to wait between successive pings",
			Value: 1 * time.Second,
		},
	},
	Action: func(c *cli.Context) {
		addr := c.GlobalString("addr")
		if addr == "" {
			addr = c.Args().First()
		}
		if addr == "" {
			fmt.Println("please provide the address to ping")
			os.Exit(1)
		}

		server, err := steam.Connect(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to server %v: %v\n", addr, err)
			os.Exit(1)
		}
		defer server.Close()

		for {
			ping, err := server.Ping()
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not ping %v: %v\n", addr, err)
				os.Exit(1)
			}

			fmt.Println(ping)

			if c.Bool("v") || c.Bool("vv") {
				info, err := server.Info()
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not get server info from %v: %v\n", addr, err)
					os.Exit(1)
				}

				fmt.Printf("\ninfo of %v:\n%v\n", addr, info)
			}

			if c.Bool("vv") {
				playersInfo, err := server.PlayersInfo()
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not get server player info from %v: %v\n", addr, err)
					os.Exit(1)
				}

				if len(playersInfo.Players) > 0 {
					fmt.Printf("\nplayer infos for %v:\n", addr)
					for _, player := range playersInfo.Players {
						fmt.Printf("%v %v\n", player.Name, player.Score)
					}
				}
			}

			if !c.Bool("forever") {
				return
			}

			time.Sleep(c.Duration("delay"))

			if c.Bool("v") || c.Bool("vv") {
				fmt.Println()
			}
		}
	},
}
