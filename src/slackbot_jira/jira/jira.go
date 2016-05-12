package jira

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "time"

    "github.com/plouc/go-jira-client"

    "slackbot_jira/config"
)

// These types are lifted straight from https://github.com/plouc/go-jira-client/blob/master/jira.go
/*
type ActivityItem struct {
    Title    string    `xml:"title"json:"title"`
    Id       string    `xml:"id"json:"id"`
    Link     []Link    `xml:"link"json:"link"`
    Updated  time.Time `xml:"updated"json:"updated"`
    Author   Person    `xml:"author"json:"author"`
    Summary  Text      `xml:"summary"json:"summary"`
    Category Category  `xml:"category"json:"category"`
}

type activityFeed struct {
    XMLName  xml.Name        `xml:"http://www.w3.org/2005/Atom feed"json:"xml_name"`
    Title    string          `xml:"title"json:"title"`
    Id       string          `xml:"id"json:"id"`
    Link     []Link          `xml:"link"json:"link"`
    Updated  time.Time       `xml:"updated,attr"json:"updated"`
    Author   Person          `xml:"author"json:"author"`
    Entries  []*ActivityItem `xml:"entry"json:"entries"`
}
*/

type Jira interface {
   GetNewActivities(time.Time) ([]*gojira.ActivityItem, error)
}

func New(cfg config.JiraConfig) Jira {
    return &jira{cfg}
}

type jira struct {
    cfg config.JiraConfig
}

func (j *jira) GetNewActivities(since int64, last_id_seen string) ([]*gojira.ActivityItem, error) {
	// TODO
	url := fmt.Sprintf("https://cera.davies:blah@learnosity.atlassian.net/activity?streams=update-date+AFTER+%d", since.Unix())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

    var activities gojira.ActivityFeed
    dec := xml.NewDecoder(resp.Body)
    err = dec.Decode(&activities)
    if err != nil {
        return nil, err
    }
    return activities.Entries, nil
}
