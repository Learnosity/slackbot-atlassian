package slack

import (
	"slackbot_jira/config"
)

type Slack interface {
	PostMessage(channel string, as_user, message string)
}

func New(config.SlackConfig) (Slack, error) {
	// TODO
	return nil, nil
}
