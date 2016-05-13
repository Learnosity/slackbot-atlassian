package message

import (
	"regexp"

	"slackbot_atlassian/atlassian"
	"slackbot_atlassian/config"
)

type Message struct {
	SlackChannel string
	AsUser       config.SlackUser
	Text         string
}

type MessageMatcher interface {
	GetMatchingMessages([]config.MessageTrigger, ...atlassian.ActivityIssue) []Message
}

type matcher struct {
	cfg config.SlackConfig
}

func NewMessageMatcher(cfg config.SlackConfig) MessageMatcher {
	return matcher{cfg}
}

func (m matcher) GetMatchingMessages(triggers []config.MessageTrigger, activity_issues ...atlassian.ActivityIssue) []Message {
	messages := make([]Message, 0)

	for _, activity_issue := range activity_issues {
		for _, trigger := range triggers {
			if match, ok := m.get_match(trigger, activity_issue); ok {
				messages = append(messages, match.get_messages()...)
			}
		}
	}

	return messages
}

func (m matcher) get_match(trigger config.MessageTrigger, activity_issue atlassian.ActivityIssue) (match, bool) {
	// TODO
	return match{trigger, activity_issue}, true
}

type match struct {
	trigger        config.MessageTrigger
	activity_issue atlassian.ActivityIssue
}

func (m match) get_messages() []Message {
	message := Message{
		m.trigger.SlackChannel,
		config.SlackUser{
			Name: m.activity_issue.Activity.Author.Name,
		},
		GetTextFromActivityItem(m.activity_issue.Activity),
	}
	return []Message{message}
}

func GetTextFromActivityItem(activity *atlassian.ActivityItem) string {
	// Strip name from start of title
	re := regexp.MustCompile("^<a.+?</a>")
	text := re.ReplaceAllString(activity.Title, "")

	// Convert HTML links
	re = regexp.MustCompile(`<a .*?href="(.+?)".*?>(.+?)</a>`)
	text = re.ReplaceAllString(text, "<$1|$2>")

	// Strip duplicate whitespace
	re = regexp.MustCompile(" +")
	text = re.ReplaceAllString(text, " ")

	return text
}
