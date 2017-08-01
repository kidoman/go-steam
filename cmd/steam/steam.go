package main

import "github.com/urfave/cli"

func main() {
	app := cli.NewApp()
	app.Name = "steam"
	app.Version = "0.1.0"
	app.Usage = "talk to source servers"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "addr",
			Usage:  "server to connect to",
			EnvVar: "ADDR",
		},
	}

	app.Commands = commands

	app.RunAndExitOnError()
}

var commands []cli.Command

func registerCommand(cmd *cli.Command) {
	commands = append(commands, *cmd)
}
