package atlassian

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"slackbot_atlassian/config"
	"slackbot_atlassian/log"
)

// These types are lifted from https://github.com/plouc/go-jira-client/blob/master/jira.go
// but with modification.

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"json:"rel"`
	Href string `xml:"href,attr"json:"href"`
}

type Person struct {
	Name     string `xml:"name"json:"name"`
	URI      string `xml:"uri"json:"uri"`
	Email    string `xml:"email"json:"email"`
	InnerXML string `xml:",innerxml"json:"inner_xml"`
}

type Text struct {
	Type string `xml:"type,attr,omitempty"json:"type"`
	Body string `xml:",chardata"json:"body"`
}

type Category struct {
	Term string `xml:"term,attr"json:"term"`
}

/*
<activity:target>
<id>urn:uuid:bebc2b8e-415d-3b15-b23d-ad1d9892ee43</id>
<title type="text">LRN-9244</title>
<summary type="text">Make 'Check Answer' button colour Accessible</summary>
<link rel="alternate" href="https://learnosity.atlassian.net/browse/LRN-9244"/>
<activity:object-type>http://streams.atlassian.com/syndication/types/issue</activity:object-type>
</activity:target>
*/
type ActivityTargetOrObject struct {
	Id         string `xml:"id"`
	Title      string `xml:"title"`
	Summary    string `xml:"summary"`
	Link       Link   `xml:"link"`
	ObjectType string `xml:"activity:object-type"`
}

/*
<activity:object>
<id>urn:uuid:be918ed4-d318-3343-8f90-aa74d24b973a</id>
<title type="text">LRN-115</title>
<summary type="text">Author API v1.0.0</summary>
<link rel="alternate" href="https://learnosity.atlassian.net/browse/LRN-115"/>
<activity:object-type>http://streams.atlassian.com/syndication/types/issue</activity:object-type>
</activity:object>
*/
/*
type ActivityObject struct {
    Id         string `xml:"id"`
    Title      string `xml:"title"`
    Summary    string `xml:"summary"`
    Link       Link   `xml:"link"`
    ObjectType string `xml:"activity:object-type"`
}
*/

type ActivityItem struct {
	Title          string                  `xml:"title"json:"title"`
	Id             string                  `xml:"id"json:"id"`
	Link           []Link                  `xml:"link"json:"link"`
	Updated        time.Time               `xml:"updated"json:"updated"`
	Author         Person                  `xml:"author"json:"author"`
	Summary        Text                    `xml:"summary"json:"summary"`
	Category       Category                `xml:"category"json:"category"`
	ActivityTarget *ActivityTargetOrObject `xml:"target"`
	ActivityObject *ActivityTargetOrObject `xml:"object"`
}

type ActivityFeed struct {
	XMLName xml.Name        `xml:"http://www.w3.org/2005/Atom feed"json:"xml_name"`
	Title   string          `xml:"title"json:"title"`
	Id      string          `xml:"id"json:"id"`
	Link    []Link          `xml:"link"json:"link"`
	Updated time.Time       `xml:"updated,attr"json:"updated"`
	Author  Person          `xml:"author"json:"author"`
	Entries []*ActivityItem `xml:"entry"json:"entries"`
}

func (ai ActivityItem) GetIssueID() (string, bool) {
	if ai.ActivityTarget != nil {
		return ai.ActivityTarget.Title, true
	}
	if ai.ActivityObject != nil {
		return ai.ActivityObject.Title, true
	}
	return "", false
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
	GetNewJiraActivities(last_id_seen string) ([]*ActivityItem, error)
	GetIssue(id string) (*Issue, error)
}

func New(cfg config.AtlassianConfig) Atlassian {
	return &atlassian{cfg}
}

type atlassian struct {
	cfg config.AtlassianConfig
}

const (
	atlassian_provider = "issues"
)

func (a *atlassian) GetNewJiraActivities(last_id_seen string) ([]*ActivityItem, error) {
	urlTemplate := "https://%s:%s@%s/activity?maxResults=%d&providers=%s"

	url := fmt.Sprintf(urlTemplate,
		a.cfg.Auth.Username, a.cfg.Auth.Password, a.cfg.Host, a.cfg.MaxActivityLookup, atlassian_provider)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var activities ActivityFeed
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&activities)
	if err != nil {
		return nil, err
	}

	// Reverse the order of the activities
	entries := filterActivities(activities.Entries, last_id_seen)

	// Only take up to the last id seen
	return reverseActivities(entries), nil
}

func reverseActivities(a []*ActivityItem) []*ActivityItem {
	//https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
	return a
}

func filterActivities(a []*ActivityItem, last_id_seen string) []*ActivityItem {
	filter := make([]*ActivityItem, 0)
	for _, item := range a {
		if item.Id == last_id_seen {
			log.LogF("Found last ID seen: %s", last_id_seen)
			break
		}
		filter = append(filter, item)
	}
	return filter
}

func (a *atlassian) GetIssue(issue_id string) (*Issue, error) {
	url := fmt.Sprintf(
		"https://%s:%s@%s/rest/api/latest/issue/%s",
		a.cfg.Auth.Username, a.cfg.Auth.Password, a.cfg.Host, issue_id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad status code looking up issue %s: %d", issue_id, resp.StatusCode)
	}

	var issue Issue
	return &issue, decodeJson(resp.Body, &issue)
}

func decodeJson(rdr io.Reader, into interface{}) error {
	dec := json.NewDecoder(rdr)
	return dec.Decode(into)
}
