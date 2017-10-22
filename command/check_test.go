package command_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gmail "google.golang.org/api/gmail/v1"

	"github.com/guywithnose/runner"
	"github.com/guywithnose/unreadChecker/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestCmdCheck(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	ts := getMockGoogleAPI(t)
	defer ts.Close()
	command.BasePath = ts.URL
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	var OAuthURL string
	ec.Closure = func(command string) {
		OAuthURL = strings.Replace(command, "xdg-open ", "", -1)
		go func() {
			_, err := http.Get(OAuthURL)
			assert.Nil(t, err)
		}()
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	app, writer, _, set := getBaseAppAndFlagSet(t, testFolder, ts.URL)
	assert.Nil(t, command.CmdCheck(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, fmt.Sprintf("Attempting to open %s in your browser\n4\n", OAuthURL), writer.String())
}

func TestCmdCheckInboxFailure(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	ts := getMockGoogleAPIInboxFailure(t)
	defer ts.Close()
	command.BasePath = ts.URL
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	var OAuthURL string
	ec.Closure = func(command string) {
		OAuthURL = strings.Replace(command, "xdg-open ", "", -1)
		go func() {
			_, err := http.Get(OAuthURL)
			assert.Nil(t, err)
		}()
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	app, _, _, set := getBaseAppAndFlagSet(t, testFolder, ts.URL)
	assert.EqualError(t, command.CmdCheck(cb)(cli.NewContext(app, set, nil)), "Unable to check inbox. googleapi: got HTTP response code 500 with body: ")
}

func TestCmdCheckTokenFailure(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	ts := getMockGoogleAPITokenFailure(t)
	defer ts.Close()
	command.BasePath = ts.URL
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	ec.Closure = func(command string) {
		go func() {
			_, err := http.Get(strings.Replace(command, "xdg-open ", "", -1))
			assert.Nil(t, err)
		}()
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	app, _, _, set := getBaseAppAndFlagSet(t, testFolder, ts.URL)
	assert.EqualError(
		t,
		command.CmdCheck(cb)(cli.NewContext(app, set, nil)),
		"Could not get OAuth token: Unable to retrieve token from web: oauth2: cannot fetch token: 500 Internal Server Error\nResponse: ",
	)
}

func TestCmdCheckUsage(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	app, _, _, set := getBaseAppAndFlagSet(t, testFolder, "")
	assert.Nil(t, set.Parse([]string{"foo"}))
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdCheck(cb)(cli.NewContext(app, set, nil)), `Usage: "unreadChecker check"`)
}

func TestCmdCheckMissingCredentialFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	set := flag.NewFlagSet("test", 0)
	tokenFile := filepath.Join(testFolder, "tokenFile")
	set.String("tokenFile", tokenFile, "doc")
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdCheck(cb)(cli.NewContext(app, set, nil)), `You must specify a credentialFile`)
}

func TestCmdCheckInvalidCredentialFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	set := flag.NewFlagSet("test", 0)
	credentialFile := filepath.Join(testFolder, "credFile")
	tokenFile := filepath.Join(testFolder, "tokenFile")
	set.String("credentialFile", credentialFile, "doc")
	set.String("tokenFile", tokenFile, "doc")
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(
		t,
		command.CmdCheck(cb)(cli.NewContext(app, set, nil)),
		`Could not initialize token client: Unable to read app credential file: open /tmp/testUnreadChecker/credFile: no such file or directory`,
	)
}

func TestCmdCheckMissingTokenFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testUnreadChecker")
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	defer removeFile(t, testFolder)
	set := flag.NewFlagSet("test", 0)
	credentialFile := filepath.Join(testFolder, "credFile")
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(""), 0777))
	set.String("credentialFile", credentialFile, "doc")
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdCheck(cb)(cli.NewContext(app, set, nil)), `You must specify a tokenFile`)
}

func getBaseAppAndFlagSet(t *testing.T, testFolder, mockAPIURL string) (*cli.App, *bytes.Buffer, *bytes.Buffer, *flag.FlagSet) {
	set := flag.NewFlagSet("test", 0)
	credentialFile := filepath.Join(testFolder, "credFile")
	tokenFile := filepath.Join(testFolder, "tokenFile")
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(mockAPIURL), 0777))
	set.String("credentialFile", credentialFile, "doc")
	set.String("tokenFile", tokenFile, "doc")
	app, writer, errorWriter := appWithTestWriters()
	return app, writer, errorWriter, set
}

func removeFile(t *testing.T, fileName string) {
	assert.Nil(t, os.RemoveAll(fileName))
}

func getTestCredentials(fakeAPI string) []byte {
	type cred struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
	}
	bytes, _ := json.Marshal(
		map[string]cred{
			"installed": {
				ClientID:     "id",
				ClientSecret: "secret",
				RedirectURIs: []string{fakeAPI},
				AuthURI:      fakeAPI,
				TokenURI:     fakeAPI,
			},
		},
	)

	return bytes
}

func getMockGoogleAPI(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		if strings.Contains(r.URL.String(), "access_type=offline") {
			go func() {
				_, err = http.Get(fmt.Sprintf("%s?code=foo", r.FormValue("redirect_uri")))
				assert.Nil(t, err)
			}()
			return
		}

		if r.URL.String() == "/me/messages?alt=json&labelIds=INBOX&q=label%3Aunread" {
			resp := gmail.ListMessagesResponse{
				Messages: []*gmail.Message{
					{},
					{},
				},
				NextPageToken: "page2",
			}
			bytes, _ := json.Marshal(resp)
			_, err = w.Write(bytes)
			assert.Nil(t, err)
			return
		}

		if r.URL.String() == "/me/messages?alt=json&labelIds=INBOX&pageToken=page2&q=label%3Aunread" {
			resp := gmail.ListMessagesResponse{
				Messages: []*gmail.Message{
					{},
					{},
				},
			}
			bytes, _ := json.Marshal(resp)
			_, err = w.Write(bytes)
			assert.Nil(t, err)
			return
		}

		require.Equal(t, "foo", r.FormValue("code"))
		response := url.Values{"access_token": []string{"fakeToken"}}
		_, err = w.Write([]byte(response.Encode()))
		assert.Nil(t, err)
	}))
}

func getMockGoogleAPIInboxFailure(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		if strings.Contains(r.URL.String(), "access_type=offline") {
			go func() {
				_, err = http.Get(fmt.Sprintf("%s?code=foo", r.FormValue("redirect_uri")))
				assert.Nil(t, err)
			}()
			return
		}

		if r.URL.String() == "/me/messages?alt=json&labelIds=INBOX&q=label%3Aunread" {
			w.WriteHeader(500)
			return
		}

		require.Equal(t, "foo", r.FormValue("code"))
		response := url.Values{"access_token": []string{"fakeToken"}}
		_, err = w.Write([]byte(response.Encode()))
		assert.Nil(t, err)
	}))
}

func getMockGoogleAPITokenFailure(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		if strings.Contains(r.URL.String(), "access_type=offline") {
			go func() {
				_, err = http.Get(fmt.Sprintf("%s?code=foo", r.FormValue("redirect_uri")))
				assert.Nil(t, err)
			}()
			return
		}

		w.WriteHeader(500)
	}))
}
