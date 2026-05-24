package main

import "github.com/task-cli/task/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
