package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

const ENV_VAR = "CONFIG"

type MessageTrigger struct {
	SlackChannel  string            `json:"slack_channel"`
	Match         map[string]string `json:"match"`
	matchCompiled map[string]*regexp.Regexp
}

func (mt MessageTrigger) GetCompiledMatches() map[string]*regexp.Regexp {
	return mt.matchCompiled
}

type StateConfig struct {
	Driver string      `json:"redis"`
	Host   string      `json:"host"`
	Port   int         `json:"port"`
	DB     interface{} `json:"db"`
}

type AtlassianConfig struct {
	Host string `json:"host"`
    MaxActivityLookup int `json:"max_activity_lookup"`
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

type SlackConfig struct {
	TeamDomain string `json:"team_domain"`
	Auth       struct {
		Token string `json:"token"`
	} `json:"auth"`
	Users map[string]SlackUser `json:"users"`
}

type SlackUser struct {
	Name      string `json:"name"`
	IconUrl   string `json:"icon_url"`
	IconEmoji string `json:"icon_emoji"`
}

type CustomJiraFieldConfig struct {
	Name      string `json:"name"`
	JiraField string `json:"jira_field"`
}

type Config struct {
	State            StateConfig             `json:"state"`
	Atlassian        AtlassianConfig         `json:"atlassian"`
	Slack            SlackConfig             `json:"slack"`
	Triggers         []MessageTrigger        `json:"triggers"`
	CustomJiraFields []CustomJiraFieldConfig `json:"custom_jira_fields"`
}

func LoadConfig(input io.Reader) (*Config, error) {
	// Parse the config JSON
	var cfg Config
	dec := json.NewDecoder(input)
	err := dec.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// Compile the match regular expression
	for _, t := range cfg.Triggers {
		t.matchCompiled = make(map[string]*regexp.Regexp)
		for k, v := range t.Match {
			match, err := regexp.Compile(v)
			if err != nil {
				return nil, fmt.Errorf("Invalid regexp %q: %s", v, err)
			}
			t.matchCompiled[k] = match
		}
	}

	return &cfg, nil
}

func LoadConfigEnv() (*Config, error) {
	filepath := os.Getenv(ENV_VAR)
	if filepath == "" {
		return nil, fmt.Errorf("Environment variable %q for config not found", ENV_VAR)
	}
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Could not open config file %q: %s", filepath, err)
	}
	defer f.Close()
	return LoadConfig(f)
}
