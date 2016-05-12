// +build integration

package slack

import (
	"testing"

	"slackbot_atlassian/config"
)

func TestSlackPostMessage(t *testing.T) {
	cfg, err := config.LoadConfigEnv()

	if err != nil {
		t.Fatal("Couldn't load config:", err)
	}

	slack := New(cfg.Slack)

	user := config.SlackUser{
		Name:      "Bob",
		IconEmoji: ":skull:",
	}

	err = slack.PostMessage("@slackbot", user, "hello world")

	if err != nil {
		t.Error(err)
	}

	err = slack.PostMessage("wrong_channel_name_kjsdfkjsf", user, "hello world")

	if err == nil {
		t.Error("Expected error for incorrect channel name")
	}
}
