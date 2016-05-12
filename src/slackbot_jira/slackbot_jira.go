package slackbot_jira

import (
    "fmt"
    "time"

    "slackbot_jira/config"
    "slackbot_jira/slack"
    "slackbot_jira/jira"
    "slackbot_jira/state"
)

// This function:
//
// * reads the last event from Redis
// * queries Jira to get new activities
// * processes each activity and posts it to Slack
func ProcessActivityStream(config *config.Config) error {
    // Get access to our state
    _, err := state.New(config.State)
    if err != nil {
        return err
    }

    // Get a Jira client
    jira := jira.New(config.Jira)

    // Get a Slack client
    _, err = slack.New(config.Slack)
    if err != nil {
        return err
    }

    // Do stuff
    since := time.Now().AddDate(0, 0, -1)
    activities, err := jira.GetNewActivities(since)
    if err != nil {
        return err
    }
    fmt.Printf("Activities since %s\n", since)
    for _, activity := range activities {
        fmt.Println(activity.Title)
    }
    return nil
}
