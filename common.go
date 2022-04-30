package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeebo/errs/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
	"log"
	"os/user"
	"path"
	"strings"
	"time"
)

type BarItem struct {
	Text    string `json:"text,omitempty"`
	Alt     string `json:"alt,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
}

func readToken(dir string) (*oauth2.Token, error) {
	t := &oauth2.Token{}
	content, err := ioutil.ReadFile(path.Join(dir, "token.json"))
	if err != nil {
		return t, errs.Wrap(err)
	}
	err = json.Unmarshal(content, t)
	if err != nil {
		return t, errs.Wrap(err)
	}
	return t, nil
}

func readCredentials(configDir string) (*oauth2.Config, error) {
	credentialFile := path.Join(configDir, "credentials.json")
	content, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		return nil, errs.Errorf("Couldn't read credentials file from %s: %v", credentialFile, err)
	}

	config, err := google.ConfigFromJSON(content, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Couldn't parse configuration: %v", err)
	}
	return config, nil
}

func setup(configDir string) (err error) {
	config, err := readCredentials(configDir)
	if err != nil {
		return errs.Wrap(err)
	}

	ctx := context.Background()
	token, _ := readToken(configDir)

	token.Expiry = time.Now().Add(-time.Hour)

	if !token.Valid() {
		if token.RefreshToken != "" {
			token, err = config.TokenSource(ctx, token).Token()
			if err != nil {
				fmt.Println(err)
			}
		}
		if !token.Valid() {
			fmt.Println(config.AuthCodeURL("no-state", oauth2.AccessTypeOffline))
			var authCode string
			if _, err := fmt.Scan(&authCode); err != nil {
				return errs.Wrap(err)
			}
			token, err := config.Exchange(ctx, authCode)
			if err != nil {
				if _, err := fmt.Scan(&authCode); err != nil {
					return errs.Wrap(err)
				}
			}
			tokenBytes, err := json.Marshal(token)
			if err != nil {
				return errs.Wrap(err)
			}
			err = ioutil.WriteFile(path.Join(configDir, "token.json"), tokenBytes, 0600)
			if err != nil {
				return errs.Wrap(err)
			}
		}

	}
	return nil

}

func getConfigDir(dir string) string {
	user, err := user.Current()
	if err != nil {
		return dir
	}
	return strings.ReplaceAll(dir, "${HOME}", user.HomeDir)
}
