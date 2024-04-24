package main

import "github.com/sx-network/sx-reporter/infra/root"

// main is the entry point for the program.
// It initializes the root command of the application
// and executes it to start the program.
func main() {
	root.NewRootCommand().Execute()
}
