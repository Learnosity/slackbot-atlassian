package config_test

import (
	"bytes"
	"strings"
	"testing"

	"slackbot_atlassian/config"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		input     string
		is_valid  bool
		err_match string
	}{
		{
			`{
                "state": {"driver": "redis", "host": "localhost", "port": 0, "db": 0},
                "jira": {
                    "host": "learnosity.atlassian.net",
                    "auth": {
                        "login": "foo",
                        "password": "bar"
                    }
                },
                "slack": {
                    "team_domain": "Learnosity"
                },
                "matches": [
                    {
                       "slack_channel": "team-yoda-jira",
                       "match": {
                            "team": "Yoda"
                       }
                    }
                ]
            }
            `, true, "",
		},
		{`jkldfj`, false, "invalid character"},
		{`{"state": 0}`, false, "cannot unmarshal number"},
		{
			`{
                "state": {"driver": "redis", "host": "localhost", "port": 0, "db": 0},
                "jira": {},
                "slack": {
                    "team_domain": "Learnosity"
                },
                "triggers": [
                    {
                       "slack_channel": "team-yoda-jira",
                       "match": {
                            "team": "("
                       }
                    }
                ]
            }
            `, false, "Invalid regexp",
		},
	}

	for _, c := range cases {
		buf := bytes.NewBuffer([]byte(c.input))
		_, err := config.LoadConfig(buf)
		if err == nil && !c.is_valid {
			t.Errorf("Expected JSON not to be valid config")
			t.Fail()
		} else if err != nil && c.is_valid {
			t.Errorf("Expected JSON to be valid config; got error %s", err)
			t.Fail()
		} else if err != nil && !strings.Contains(err.Error(), c.err_match) {
			t.Errorf("Expected error to contain %q, got %q", c.err_match, err)
			t.Fail()
		}
	}
}
