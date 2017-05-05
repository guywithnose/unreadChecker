package command_test

import (
	"flag"
	"os"
	"testing"

	"github.com/guywithnose/unreadChecker/command"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCompleteCheck(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	os.Args = []string{os.Args[0], "check", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "--credentialFile\n--tokenFile\n", writer.String())
}

func TestCompleteCheckCredentialFile(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	os.Args = []string{os.Args[0], "check", "--credentialFile", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestCompleteCheckTokenFile(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	os.Args = []string{os.Args[0], "check", "--tokenFile", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestRootCompletion(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = append(command.Commands, cli.Command{Hidden: true, Name: "don't show"})
	command.RootCompletion(cli.NewContext(app, set, nil))
	assert.Equal(t, "check:Check your inbox for unread messages\n", writer.String())
}
