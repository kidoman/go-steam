package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sostronk/go-steam"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	registerCommand(&rconCommand)
}

var rconCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command on server",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "forever, f",
			Usage: "exec forever",
		},
		cli.DurationFlag{
			Name:  "delay, d",
			Usage: "time to wait between successive execs",
			Value: 1 * time.Second,
		},
	},
	Action: func(c *cli.Context) {
		addr := c.GlobalString("addr")
		if addr == "" {
			fmt.Println("please provide the address to exec command on")
			os.Exit(1)
		}

		if !c.Args().Present() {
			fmt.Println("please provide the command to exec")
			os.Exit(1)
		}

		cmd := c.Args().First()

		fmt.Printf("Password: ")
		password, err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not read password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()

		server, err := steam.Connect(addr, steam.WithRCONPassword(string(password)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to server %v: %v\n", addr, err)
			os.Exit(1)
		}
		defer server.Close()

		for {
			resp, err := server.Send(cmd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not exec %q on %v: %v\n", cmd, addr, err)
				os.Exit(1)
			}

			fmt.Println(resp)

			if !c.Bool("forever") {
				return
			}

			time.Sleep(c.Duration("delay"))
		}
	},
}
