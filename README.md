# Slackbot-Atlassian

[![Build Status](https://travis-ci.org/Learnosity/slackbot-atlassian.svg?branch=master)](https://travis-ci.org/Learnosity/slackbot-atlassian)

Slackbot-Jira is a bot for [Slack](https://slack.com) that posts messages from a [Jira](https://www.atlassian.com/software/jira) activity feed.

## Usage

You need [go](https://golang.org) and [gb](https://getgb.io) install to build
the project.

After cloning, you can build and run `slackbot-atlassian`, setting an environment variable to
point it at its config file:

```bash
$ gb build all
$ CONFIG=slackbot-config.json ./bin/slackbot-atlassian
```

## Configuration

The config should be a JSON file whose structure corresponds to the `Config`
type in the code (see `src/slackbot-atlassian/config.go`).

The bot process is told where to get its config with the `CONFIG` environment
variable.

## Testing

To run the tests:

	gb test all

To include the integration tests:

    gb test -tags integration all

The integration tests assume that you have a Redis instance at `localhost:6379`.
## License

Permissively MIT-licensed. See the LICENSE file.
