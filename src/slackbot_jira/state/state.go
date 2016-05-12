package state

import (
	"encoding/json"
	"fmt"
	"time"

	"slackbot_jira/config"

	"gopkg.in/redis.v3"
)

const redis_state_key = "slackbot_jira_last_event"

type Event struct {
	TimestampSecs int64  `json:"timestamp_secs"`
	Id            string `json:"id"`
}

type State interface {
	RecordLastEvent(Event) error
	GetLastEvent() (Event, bool, error)
}

func New(cfg config.StateConfig) (State, error) {
	return new(cfg, redis_state_key)
}

func new(cfg config.StateConfig, key string) (State, error) {
	var db int64
	switch cfg.DB.(type) {
	case int:
		db = int64(cfg.DB.(int))
	case int64:
		db = cfg.DB.(int64)
	case float64:
		db = int64(cfg.DB.(float64))
	}
	redisOptions := redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DB:   db,
	}
	client := redis.NewClient(&redisOptions)
	return &redisState{client, key}, nil
}

type redisState struct {
	client *redis.Client
	key    string
}

func (r *redisState) RecordLastEvent(ev Event) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	sc := r.client.Set(r.key, string(b), time.Duration(0))
	return sc.Err()
}

func (r *redisState) GetLastEvent() (Event, bool, error) {
	var ev Event
	sc := r.client.Get(r.key)
	err := sc.Err()
	if err != nil && err == redis.Nil {
		// No key found
		return ev, false, nil
	} else if err != nil {
		// Error looking up key
		return ev, false, err
	}

	val, err := sc.Result()
	if err != nil {
		return ev, false, err
	}

	return ev, true, json.Unmarshal([]byte(val), &ev)
}
