// +build integration

package slack

import (
	"testing"

	"slackbot_jira/config"
)

func TestSlackPostMessage(t *testing.T) {
	cfg, err := config.LoadConfigEnv()

	if err != nil {
		t.Fatal("Couldn't load config:", err)
	}

	slack := New(cfg.Slack)

	user := User{
		Name:      "Bob",
		IconEmoji: ":skull:",
	}

	err = slack.PostMessage("@slackbot", user, "hello world")

	if err != nil {
		t.Fatal(err)
	}
}
