package message

import (
	"fmt"
	"regexp"

	"slackbot_atlassian/atlassian"
	"slackbot_atlassian/config"
	"slackbot_atlassian/log"
)

type Message struct {
	SlackChannel string
	AsUser       config.SlackUser
	Text         string
}

type MessageMatcher interface {
	GetMatchingMessages([]*config.MessageTrigger, ...atlassian.ActivityIssue) []Message
}

type matcher struct {
	cfg                config.SlackConfig
	custom_jira_fields []config.CustomJiraFieldConfig
	user_image_urls    map[string]string
}

func NewMessageMatcher(cfg config.SlackConfig, user_image_urls map[string]string, custom_jira_fields ...config.CustomJiraFieldConfig) MessageMatcher {
	return matcher{cfg, custom_jira_fields, user_image_urls}
}

func (m matcher) GetMatchingMessages(triggers []*config.MessageTrigger, activity_issues ...atlassian.ActivityIssue) []Message {
	messages := make([]Message, 0)

	for _, activity_issue := range activity_issues {
		for _, trigger := range triggers {
			if match, ok, err := m.get_match(trigger, activity_issue); ok && err == nil {
				messages = append(messages, match.get_messages()...)
			} else if err != nil {
				log.LogF("Error matching issue %v: %s", activity_issue, err)
			}
		}
	}

	return messages
}

func (m matcher) get_match(trigger *config.MessageTrigger, activity_issue atlassian.ActivityIssue) (*match, bool, error) {
	for name, match := range trigger.GetCompiledMatches() {
		// Look up the value for this field
		field_val, ok, err := m.get_trigger_field_value(name, activity_issue)
		if err != nil || !ok {
			return nil, ok, err
		}

		if !match.MatchString(field_val) {
			return nil, false, nil
		}
	}

	return &match{m.user_image_urls, trigger, activity_issue}, true, nil
}

func (m matcher) get_trigger_field_value(name string, activity_issue atlassian.ActivityIssue) (string, bool, error) {
	// First, check if this is a custom field defined by the JSON
	for _, cf := range m.custom_jira_fields {
		if cf.Name == name {
			val, ok := activity_issue.Issue.Fields[cf.JiraField]
			if !ok {
				return "", false, nil
			}
			field, ok := val.(map[string]interface{})
			if !ok {
				return "", false, fmt.Errorf("Wrong type for %s / %s: want string, have %T", cf.Name, cf.JiraField, val)
			}
			value := field["value"].(string)
			return value, ok, nil
		}
	}

	// Try to get this as a field from the issue
	if val, ok := activity_issue.Issue.Fields[name]; ok {
		s, ok := val.(string)
		if !ok {
			return "", false, fmt.Errorf("Wrong type for %s: want string, have %T", name, val)
		}
		return s, ok, nil
	}

	return "", false, nil
}

type match struct {
	user_image_urls map[string]string
	trigger         *config.MessageTrigger
	activity_issue  atlassian.ActivityIssue
}

func (m match) get_messages() []Message {
	message := Message{
		m.trigger.SlackChannel,
		config.SlackUser{
			Name:    m.activity_issue.Activity.Author.Name,
			IconUrl: m.user_image_urls[m.activity_issue.Activity.Author.Username],
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
	re = regexp.MustCompile("\\s+")
	text = re.ReplaceAllString(text, " ")

	return text
}
