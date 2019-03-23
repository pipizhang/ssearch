package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const APP_NAME = "SimpleSearch"

func main() {

	if len(os.Args) == 0 || (len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("usage: %s <dir>\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	Root = os.Args[1]
	if !dirExist(Root) {
		fmt.Printf("Dir `%s` does not exist", Root)
		os.Exit(0)
	}

	app := NewApp()
	app.Start()

}
