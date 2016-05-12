package message

import (
	"slackbot_jira/atlassian"
	"slackbot_jira/config"
)

type Message struct {
	Slack_Channel string
	As_User       string
	Content       string
}

func GetMatchingMessages(matches []config.MessageMatch, activity_issues ...atlassian.ActivityIssue) []Message {
	// TODO
	return nil
}
