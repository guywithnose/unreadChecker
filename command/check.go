package command

import (
	"fmt"

	gmail "google.golang.org/api/gmail/v1"

	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
)

// BasePath allows overriding the gmail API base path for testing
var BasePath string

// CmdCheck checks the inbox for unread messages
func CmdCheck(cmdBuilder runner.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != 0 {
			return cli.NewExitError("Usage: \"unreadChecker check\"", 1)
		}

		err := checkFlags(c)
		if err != nil {
			return err
		}

		tokenClient, err := NewClient(c.String("credentialFile"), c.String("tokenFile"), cmdBuilder)
		if err != nil {
			return fmt.Errorf("Could not initialize token client: %v", err)
		}

		httpClient, err := tokenClient.GetHTTPClient(c.App.Writer)
		if err != nil {
			return fmt.Errorf("Could not get OAuth token: %v", err)
		}

		srv, _ := gmail.New(httpClient)
		if BasePath != "" {
			srv.BasePath = BasePath
		}

		user := "me"

		total := 0

		resp, err := srv.Users.Messages.List(user).LabelIds("INBOX").Q("label:unread").Do()
		if err != nil {
			return fmt.Errorf("Unable to check inbox. %v", err)
		}

		total += len(resp.Messages)

		nextPageToken := resp.NextPageToken

		for nextPageToken != "" {
			resp, err = srv.Users.Messages.List(user).LabelIds("INBOX").Q("label:unread").PageToken(nextPageToken).Do()
			if err != nil {
				return fmt.Errorf("Unable to check inbox. %v", err)
			}

			total += len(resp.Messages)

			nextPageToken = resp.NextPageToken
		}

		fmt.Fprintf(c.App.Writer, "%d\n", total)
		return nil
	}
}

func checkFlags(c *cli.Context) error {
	if c.String("credentialFile") == "" {
		return cli.NewExitError("You must specify a credentialFile", 1)
	}

	if c.String("tokenFile") == "" {
		return cli.NewExitError("You must specify a tokenFile", 1)
	}

	return nil
}
