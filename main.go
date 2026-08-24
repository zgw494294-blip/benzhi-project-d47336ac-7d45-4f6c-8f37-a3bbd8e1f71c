package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/benzhi/city-tree-release/cmd/server"
)

func main() {
	if err := server.Run(os.Args[1:]); err != nil {
		if err == http.ErrServerClosed {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
