package jira

import (
	"encoding/xml"
	"fmt"
	"net/http"

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
	GetNewActivities(unix_time int64, last_id_seen string) ([]*gojira.ActivityItem, error)
    GetIssue(id string) (*gojira.Issue, error)
}

func New(cfg config.JiraConfig) Jira {
	return &jira{cfg}
}

type jira struct {
	cfg config.JiraConfig
}

func (j *jira) GetNewActivities(since int64, last_id_seen string) ([]*gojira.ActivityItem, error) {
	url := fmt.Sprintf("https://%s:%s@%s/activity?streams=update-date+AFTER+%d",
        j.cfg.Auth.Username, j.cfg.Auth.Password, j.cfg.Host, since)
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


func (j *jira) GetIssue(id string) (*gojira.Issue, error) {
    // TODO
    return nil, nil
}
