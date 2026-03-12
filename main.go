package main

import "github.com/sebrandon1/go-skylight/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	_ = cmd.Execute()
}
