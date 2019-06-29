package main

import (
	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/src/server"
)

const port = 8080

func main() {
	var err error

	// Start the server
	clog.Info("Initializing the server...")
	err = server.StartServer("", port)
	if err != nil {
		clog.FatalErr(err)
	}

}
