/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package main

import "opslevel-agent/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
