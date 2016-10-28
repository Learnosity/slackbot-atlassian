package message

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"slackbot_atlassian/atlassian"
	"slackbot_atlassian/config"
	"slackbot_atlassian/log"
)

type Message struct {
	SlackChannel string
	AsUser       config.SlackUser
	Text         string
}

func (m Message) HtmlUnescapedText() string {
	return html.UnescapeString(m.Text)
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

	lookup_field := func(name string) (string, bool, error) {
		val, ok := activity_issue.Issue.Fields[name]
		if !ok {
			return "", false, nil
		} else if val == nil {
			return "", false, nil
		}
		switch vT := val.(type) {
		case map[string]interface{}:
			return lookup_field_value_from_map(name, vT)
		case string:
			return vT, true, nil
		case []interface{}:
			// For now we just join them all with strings
			return format_field_value_from_slice(vT), true, nil
		default:
			return "", false, fmt.Errorf("Wrong type for %s: want map or string, have %T", name, val)
		}
	}

	// First try to look up for each of our custom fields
	for _, cf := range m.custom_jira_fields {
		if cf.Name == name {
			val, ok, err := lookup_field(cf.JiraField)
			if ok {
				return val, ok, err
			}
		}
	}

	// Now try with built-in fields
	return lookup_field(name)
}

// get the field value from a generic map
//
// many jira fields are structured as an object, but different fields use
// different keys for the value we care about.
func lookup_field_value_from_map(name string, m map[string]interface{}) (string, bool, error) {
	candidateKeys := []string{"value", "name"}

	for _, k := range candidateKeys {
		if v, ok := m[k]; ok {
			// We have a value - try to convert it to string
			vs, ok := v.(string)
			if !ok {
				return "", false, fmt.Errorf("Wrong type for val inside %s: want string, have %T", name, v)
			}
			return vs, true, nil
		}
	}

	return "", false, nil
}

// some field values are JSON arrays, but we just want to work with them as a
// string to keep the matching simple - so we join them up as strings with a
// comma..
func format_field_value_from_slice(s []interface{}) string {
	var strs []string
	for _, v := range s {
		strs = append(strs, fmt.Sprintf("%s", v))
	}
	return strings.Join(strs, ",")
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

	// Convert resolved links
	re = regexp.MustCompile(`<span class='resolved-link'>(.+?)</span>`)
	text = re.ReplaceAllString(text, "~$1~")

	// Convert HTML links
	re = regexp.MustCompile(`<a .*?href="(.+?)".*?>(.+?)</a>`)
	text = re.ReplaceAllString(text, "<$1|$2>")

	// Strip duplicate whitespace
	re = regexp.MustCompile("\\s+")
	text = re.ReplaceAllString(text, " ")

	return text
}
