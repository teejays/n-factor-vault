package main

import (
	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/library/env"
	"github.com/teejays/n-factor-vault/backend/library/orm"

	"github.com/teejays/n-factor-vault/backend/src/secret"
	"github.com/teejays/n-factor-vault/backend/src/server"
	"github.com/teejays/n-factor-vault/backend/src/totp"
	"github.com/teejays/n-factor-vault/backend/src/user"
	"github.com/teejays/n-factor-vault/backend/src/vault"
)

const port = 8080

func main() {
	err := mainWithError()
	if err != nil {
		clog.FatalErr(err)
	}
}

func mainWithError() error {
	var err error

	// Set the log level
	clog.LogLevel = 8
	if env.GetEnv() == env.DEV {
		clog.LogLevel = 0
	}

	// Initialize the ORM package
	err = orm.Init()
	if err != nil {
		return err
	}

	// Initialize the services: the order should be important ideally, so dependent services are initialized later
	err = user.Init()
	if err != nil {
		return err
	}

	err = vault.Init()
	if err != nil {
		return err
	}

	err = secret.Init()
	if err != nil {
		return err
	}

	err = totp.Init()
	if err != nil {
		return err
	}

	// Start the webserver
	clog.Info("Initializing the server...")
	err = server.StartServer("", port)
	if err != nil {
		return err
	}

	return nil
}
