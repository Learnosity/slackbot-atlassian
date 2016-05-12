// +build integration

package state

import (
	"testing"
	"time"

	"slackbot_jira/config"
)

func test_state(t *testing.T) State {
	cfg := config.StateConfig{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	}
	key := "state_integration_test_key"
	s, err := new(cfg, key)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return s
}

func TestState(t *testing.T) {
	s := test_state(t)

	ev := Event{time.Now().AddDate(0, -1, 0).Unix(), "blah blah"}

	err := s.RecordLastEvent(ev)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	ev2, ok, err := s.GetLastEvent()
	if err != nil {
		t.Error(err)
		t.FailNow()
	} else if !ok {
		t.Errorf("No last event found")
		t.FailNow()
	} else if ev2.TimestampSecs != ev.TimestampSecs {
		t.Errorf("Timestamps do not match")
		t.FailNow()
	} else if ev2.Id != ev.Id {
		t.Errorf("Ids do not match")
		t.FailNow()
	}
}
