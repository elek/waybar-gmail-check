package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs/v2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"
	"os"
)

func main() {
	cmd := cobra.Command{}
	configDir := cmd.PersistentFlags().String("config-dir", "${HOME}/.config/waybar-gmail-check", "Directory to store the tokens (and credentials)")
	{
		subCmd := cobra.Command{
			Use:   "run",
			Short: "Check gmail inbox and return the unread information in waybar format.",
		}
		subCmd.RunE = func(cmd *cobra.Command, args []string) error {
			return run(getConfigDir(*configDir))
		}
		cmd.AddCommand(&subCmd)
	}
	{
		subCmd := cobra.Command{
			Use:   "setup",
			Short: "Setup credentials",
		}
		subCmd.RunE = func(cmd *cobra.Command, args []string) error {
			return setup(getConfigDir(*configDir))
		}
		cmd.AddCommand(&subCmd)
	}
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("%++v", err)
	}
}

func run(configDir string) (err error) {
	ctx := context.Background()

	config, err := readCredentials(configDir)
	if err != nil {
		return errs.Wrap(err)
	}
	token, err := readToken(configDir)
	if err != nil {
		return err
	}

	gmailService, err := gmail.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return errs.Wrap(err)
	}

	user := "me"

	r, err := gmailService.Users.Messages.List(user).Q("label:inbox is:unread").Do()
	if err != nil {
		log.Fatalf("Couldn't retrieve labels: %v", err)
	}

	unreadCount := len(r.Messages)

	if unreadCount == 0 {
		return json.NewEncoder(os.Stdout).Encode(BarItem{
			Text:    "",
			Alt:     "mail",
			Tooltip: "No new mail",
		})
	}

	return json.NewEncoder(os.Stdout).Encode(BarItem{
		Text:    fmt.Sprintf("%d ✉️", unreadCount),
		Alt:     "mail",
		Tooltip: fmt.Sprintf("%d unread mail", unreadCount),
	})
}
