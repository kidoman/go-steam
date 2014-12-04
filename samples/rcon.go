package main

import "github.com/kidoman/go-steam"

func main() {
	o := &steam.ConnectOptions{RCONPassword: "password"}
	rcon, err := steam.Connect("1.2.3.4:5", o)
	if err != nil {
		panic(err)
	}
	defer rcon.Close()
	rcon.Send("status")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
