package main

import (
	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/src/env"
	"github.com/teejays/n-factor-vault/backend/src/server"
)

const port = 8080

func main() {
	var err error

	// Set the log level
	clog.LogLevel = 2
	if env.GetEnv() == env.DEV {
		clog.LogLevel = 0
	}

	// Start the server
	clog.Info("Initializing the server...")
	err = server.StartServer("", port)
	if err != nil {
		clog.FatalErr(err)
	}

}
