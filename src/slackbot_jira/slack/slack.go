package slack

import (
	"slackbot_jira/config"

	"github.com/nlopes/slack"
)

type Slack interface {
	PostMessage(channel string, as_user User, message string) error
}

type User struct {
	Name      string
	IconUrl   string
	IconEmoji string
}

type impl struct {
	cfg config.SlackConfig
	client *slack.Client
}

func New(cfg config.SlackConfig) Slack {
	return impl{cfg, slack.New(cfg.Auth.Token)}
}

func (s impl) PostMessage(channel string, user User, message string) error {
	params := slack.NewPostMessageParameters()

	params.Username = user.Name
	params.IconURL = user.IconUrl
	params.IconEmoji = user.IconEmoji

	_, _, err := s.client.PostMessage(channel, message, params)

	return err
}
