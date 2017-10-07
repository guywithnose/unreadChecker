package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/guywithnose/runner"
	"github.com/guywithnose/unreadChecker/command"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = command.Name
	app.Version = fmt.Sprintf("%s-%s", command.Version, runtime.Version())
	app.Author = "Robert Bittle"
	app.Email = "guywithnose@gmail.com"
	app.Usage = "unreadChecker check"

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
		os.Exit(2)
	}

	app.Action = command.CmdCheck(runner.Real{})
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "credentialFile",
			Usage:  "The Gmail OAuth credential file",
			EnvVar: "GMAIL_OAUTH_CREDENTIAL_FILE",
		},
		cli.StringFlag{
			Name:  "tokenFile",
			Usage: "The token file",
		},
	}
	app.ErrWriter = os.Stderr

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
