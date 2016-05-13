// +build integration

package atlassian

import (
	"testing"

	"slackbot_atlassian/config"
)

func TestSlackPostMessage(t *testing.T) {
	cfg, err := config.LoadConfigEnv()

	if err != nil {
		t.Fatal("Couldn't load config:", err)
	}

	atl := New(cfg.Atlassian)

	cases := []struct {
		issue string
		valid bool
	}{
		{"LRN-8770", true},
		{"LRN-99990", false},
	}

	for _, c := range cases {
		_, err := atl.GetIssue(c.issue)
		if err != nil && c.valid {
			t.Fatal(err)
		} else if err == nil && !c.valid {
			t.Fatal("Issue should not exist")
		}
	}
}
