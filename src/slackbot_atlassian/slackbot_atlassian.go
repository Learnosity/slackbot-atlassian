package slackbot_atlassian

import (
	"slackbot_atlassian/atlassian"
	"slackbot_atlassian/config"
	"slackbot_atlassian/log"
	"slackbot_atlassian/message"
	"slackbot_atlassian/slack"
	"slackbot_atlassian/state"
	"slackbot_atlassian/storage"
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

	log.LogF("Creating a storage (S3) client")
	storage_client := storage.New(config.ResourceStorage)

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

	activity_issues := get_issues(config, atl, activities)

	user_image_urls := get_user_image_urls(storage_client, atl, s, activity_issues)

	matcher := message.NewMessageMatcher(config.Slack, user_image_urls, config.CustomJiraFields...)

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

func get_issues(config *config.Config, atl atlassian.Atlassian, activities []*atlassian.ActivityItem) []atlassian.ActivityIssue {
	activity_issues := make([]atlassian.ActivityIssue, 0)

	type result struct {
		*atlassian.ActivityIssue
		Ok bool
	}

	// Create a buffered channel with all the work to be done and fill it up
	input := make(chan *atlassian.ActivityItem, len(activities))
	for _, activity := range activities {
		input <- activity
	}
	defer close(input)

	// Create a buffered channel for the work results
	output := make(chan result, len(activities))
	defer close(output)

	// Create our worker goroutines
	for i := 0; i < config.Atlassian.ConcurrentIssueLookups; i++ {
		go func() {
			for activity := range input {
				issue_id, ok := activity.GetIssueID()
				if !ok {
					log.LogF("Could not get issue ID off activity")
					output <- result{nil, false}
					continue
				}

				issue, err := atl.GetIssue(issue_id)
				if err != nil {
					log.LogF("Could not find issue %s - %s", issue, err)
					output <- result{nil, false}
					continue
				}

				output <- result{&atlassian.ActivityIssue{activity, issue}, true}
			}
		}()
	}

	// Collect the output
	for i := 0; i < len(activities); i++ {
		res := <-output
		if res.Ok {
			activity_issues = append(activity_issues, *res.ActivityIssue)
		}
	}

	return activity_issues
}

func get_user_image_urls(storage_client storage.Client, atlassian_client atlassian.Atlassian, state_client state.State, activity_issues []atlassian.ActivityIssue) map[string]string {
	urls := make(map[string]string)
	for _, ai := range activity_issues {
		name := ai.Activity.Author.Username

		_, ok := urls[name]
		if ok {
			continue
		}

		url, ok, err := state_client.GetUserImageURL(name)
		if ok && err == nil {
			urls[name] = url
			continue
		} else if err != nil {
			log.LogF("Could not retrieve URL for user %s in Redis: %s", name, err)
		}

		// Retrieve and record it instead
		rdr, ok, err := atlassian_client.UserImage(*ai.Activity)
		if err != nil {
			log.LogF("Could not retrieve image for user %s from Jira: %s", name, err)
			continue
		} else if !ok {
			log.LogF("Could not find any image for user %s in Redis", name)
			continue
		}

		path := "users/images/" + name
		err = storage_client.PutObject(rdr, path)
		if err != nil {
			log.LogF("Failed to persist image for user %s: %s", name, err)
			continue
		}

		full_url := storage_client.GetFullURL(path)
		urls[name] = full_url

		// Cache it for next time
		err = state_client.RecordUserImageURL(name, full_url)
		if err != nil {
			log.LogF("Failed to save image URL for user %s: %s", name, err)
		}

	}

	return urls
}
