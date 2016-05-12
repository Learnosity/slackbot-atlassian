package slackbot_jira

import (
	"fmt"
	"os"
	"time"

	"slackbot_jira/atlassian"
	"slackbot_jira/config"
	"slackbot_jira/message"
	"slackbot_jira/slack"
	"slackbot_jira/state"
)

func logF(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

// This function:
//
// * reads the last event from Redis
// * queries Jira to get new activities
// * processes each activity and posts it to Slack
func ProcessActivityStream(config *config.Config) error {
	// Get access to our state
	logF("Creating Redis client")
	s, err := state.New(config.State)
	if err != nil {
		return err
	}

	logF("Creating jira client")
	// Get a Jira client
	atl := atlassian.New(config.Atlassian)

	logF("Creating slack client")
	// Get a Slack client
	slack_client := slack.New(config.Slack)

	logF("Looking for last event")
	// Get the last event
	lastEvent, ok, err := s.GetLastEvent()
	if err != nil {
		return err
	}
	if !ok {
		logF("No last event found - starting from now")
		// Fabricate an event
		lastEvent = state.Event{time.Now().Unix(), ""}
	}

	// Get activities since this event
	activities, err := atl.GetNewJiraActivities(lastEvent.TimestampSecs-1, lastEvent.Id)
	if err != nil {
		return err
	}

	logF("Found %d activities since %d\n", len(activities), lastEvent.TimestampSecs)

	activity_issues := make([]atlassian.ActivityIssue, 0)

	// Loop backwards so that we go from oldest to newest
	for _, activity := range activities {
		// Look up the issue
		issue_id, err := activity.GetIssueID()
		if err != nil {
			logF("Could not get issue ID off activity: %s", err)
			continue
		}

		issue, err := atl.GetIssue(issue_id)
		if err != nil {
			logF("Could not find issue %s - %s", issue, err)
			continue
		}

		activity_issues = append(activity_issues, atlassian.ActivityIssue{activity, issue})
	}
	fmt.Println(activity_issues)
	matcher := message.NewMessageMatcher(config.Slack)

	messages := matcher.GetMatchingMessages(config.Triggers, activity_issues...)

	for _, m := range messages {
		if err := slack_client.PostMessage(m.SlackChannel, m.AsUser, m.Text); err != nil {
			return err
		}
	}

	if len(activities) > 0 {
		// Record the last event seen
		lastEvent := state.Event{activities[0].Updated.Unix(), activities[0].Id}
		err := s.RecordLastEvent(lastEvent)
		if err != nil {
			return err
		}
	}

	return nil
}
