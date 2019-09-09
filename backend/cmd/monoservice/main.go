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

	clog.Info("Starting the monoservice...")

	// Set the log level
	clog.LogLevel = 2
	clog.Infof("Application Environment: %s", env.GetAppEnv())
	if env.GetAppEnv() == env.DEV {
		clog.LogLevel = 0
	}

	// Initialize the ORM package
	clog.Info("Initializing ORM...")
	err = orm.Init()
	if err != nil {
		return err
	}

	// Initialize the services: the order should be important ideally, so dependent services are initialized later
	clog.Info("Initializing User Service...")
	err = user.Init()
	if err != nil {
		return err
	}

	clog.Info("Initializing Vault Service...")
	err = vault.Init()
	if err != nil {
		return err
	}

	clog.Info("Initializing Secret Service...")
	err = secret.Init()
	if err != nil {
		return err
	}

	clog.Info("Initializing TOTP Service...")
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
