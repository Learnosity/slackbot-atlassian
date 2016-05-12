package config

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "regexp"
)

const ENV_VAR = "CONFIG"

type MessageMatch struct {
    SlackChannel string `json:"slack_channel"`
    Match map[string]string `json:"match"`
    matchCompiled map[string]*regexp.Regexp
}

type StateConfig struct {
    Driver string `json:"redis"`
    Host string `json:"host"`
    Port int `json:"port"`
    DB interface{} `json:"db"`
}

type JiraConfig struct {
    Host string `json:"base_url"`
    Auth struct {
        User string `json:"user"`
        Password string `json:"password"`
    } `json:"auth"`
}

type SlackConfig struct {
    TeamDomain string `json:"team_domain"`
    Auth struct {
        Token string `json:"token"`
    } `json:"auth"`
}

type Config struct {
    State StateConfig `json:"state"`
    Jira JiraConfig `json:"jira"`
    Slack SlackConfig `json:"slack"`
    Matches []MessageMatch `json:"matches"`
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
    for _, m := range cfg.Matches {
        m.matchCompiled = make(map[string]*regexp.Regexp)
        for k, v := range m.Match {
            match, err := regexp.Compile(v)
            if err != nil {
                return nil, fmt.Errorf("Invalid regexp %q: %s", v, err)
            }
            m.matchCompiled[k] = match
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
