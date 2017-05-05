package command

import (
	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
)

// Commands defines the commands that can be called on hostBuilder
var Commands = []cli.Command{
	{
		Name:         "check",
		Aliases:      []string{"c"},
		Usage:        "Check your inbox for unread messages",
		Action:       CmdCheck(runner.Real{}),
		BashComplete: Completion,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "credentialFile",
				Usage:  "The Gmail OAuth credential file",
				EnvVar: "GMAIL_OAUTH_CREDENTIAL_FILE",
			},
			cli.StringFlag{
				Name:  "tokenFile",
				Usage: "The token file",
			},
		},
	},
}
