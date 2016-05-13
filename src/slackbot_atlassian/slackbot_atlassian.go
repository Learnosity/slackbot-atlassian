package slackbot_atlassian

import (
	"slackbot_atlassian/atlassian"
	"slackbot_atlassian/config"
	"slackbot_atlassian/log"
	"slackbot_atlassian/message"
	"slackbot_atlassian/slack"
	"slackbot_atlassian/state"
)

// This function:
//
// * reads the last event from Redis
// * queries Jira to get new activities
// * processes each activity and posts it to Slack
func ProcessActivityStream(config *config.Config) error {
	// Get access to our state
	log.LogF("Creating Redis client")
	s, err := state.New(config.State)
	if err != nil {
		return err
	}

	log.LogF("Creating jira client")
	// Get a Jira client
	atl := atlassian.New(config.Atlassian)

	log.LogF("Creating slack client")
	// Get a Slack client
	slack_client := slack.New(config.Slack)

	log.LogF("Looking for last event")
	// Get the last event
	lastEvent, ok, err := s.GetLastEvent()
	if err != nil {
		return err
	}
	if !ok {
		log.LogF("No last event found - starting from now")
		// Fabricate an event
		lastEvent = state.Event{""}
	}

	// Get activities since this event
	activities, err := atl.GetNewJiraActivities(lastEvent.Id)
	if err != nil {
		return err
	}

	log.LogF("Found %d new activities since last event %v", len(activities), lastEvent)

	activity_issues := make([]atlassian.ActivityIssue, 0)
	// Loop backwards so that we go from oldest to newest
	for _, activity := range activities {
		// Look up the issue
		issue_id, ok := activity.GetIssueID()
		if !ok {
			log.LogF("Could not get issue ID off activity: %s", err)
			continue
		}

		issue, err := atl.GetIssue(issue_id)
		if err != nil {
			log.LogF("Could not find issue %s - %s", issue, err)
			continue
		}

		activity_issues = append(activity_issues, atlassian.ActivityIssue{activity, issue})
	}

	matcher := message.NewMessageMatcher(config.Slack, config.CustomJiraFields...)

	messages := matcher.GetMatchingMessages(config.Triggers, activity_issues...)

	log.LogF("Posting %d messages to Slack", len(messages))
	for _, m := range messages {
		if err := slack_client.PostMessage(m.SlackChannel, m.AsUser, m.Text); err != nil {
			return err
		}
	}

	if len(activities) > 0 {
		// Record the last event seen
		lastEvent = state.Event{activities[len(activities)-1].Id}
	}

	log.LogF("Record last event in state DB: %v", lastEvent)
	err = s.RecordLastEvent(lastEvent)
	if err != nil {
		return err
	}

	return nil
}
