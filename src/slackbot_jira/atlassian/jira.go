package atlassian

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

type ActivityItem gojira.ActivityItem

func (ai ActivityItem) GetIssueID() (string, error) {
	// TODO
	return "", nil
}

type Issue struct {
	Id     string `json:"key"`
	Fields map[string]interface{}
}

type ActivityIssue struct {
	Activity *ActivityItem
	Issue    *Issue
}

type Atlassian interface {
	GetNewJiraActivities(unix_time int64, last_id_seen string) ([]*ActivityItem, error)
	GetIssue(id string) (*Issue, error)
}

func New(cfg config.AtlassianConfig) Atlassian {
	auth := gojira.Auth{Login: cfg.Auth.Username, Password: cfg.Auth.Password}
	jira := gojira.NewJira("https://"+cfg.Host, "/rest/api/2", "/activity", &auth)
	return &atlassian{cfg, jira}
}

type atlassian struct {
	cfg        config.AtlassianConfig
	jiraClient *gojira.Jira
}

const (
	atlassian_provider = "issues"
)

func (a *atlassian) GetNewJiraActivities(since int64, last_id_seen string) ([]*ActivityItem, error) {
	url := fmt.Sprintf(
		"https://%s:%s@%s/activity?providers=%s&streams=update-date+AFTER+%d",
		a.cfg.Auth.Username, a.cfg.Auth.Password, a.cfg.Host, atlassian_provider, since)
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

	// Convert to our own type
	entries := make([]*ActivityItem, len(activities.Entries))
	for i, a := range activities.Entries {
		item := *a
		act := ActivityItem(item)
		entries[i] = &act
	}
	return entries, nil
}

func (a *atlassian) GetIssue(id string) (*Issue, error) {
	// TODO
	return nil, nil
}
