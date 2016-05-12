package main

import (
	"fmt"
	"os"

	"slackbot_atlassian"
	"slackbot_atlassian/config"
)

func failF(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func main() {
	cfg, err := config.LoadConfigEnv()
	if err != nil {
		failF("Failed to load config: %s", err)
		os.Exit(1)
	}

	err = slackbot_atlassian.ProcessActivityStream(cfg)
	if err != nil {
		failF("Error while processing activity stream: %s\n", err)
		os.Exit(1)
	}
}
