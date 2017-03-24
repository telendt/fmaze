package main

import (
	"errors"
	"fmt"
)

// ErrEventParseError is returned by ParseEvent to indicate that problems with
// parsing event's payload.
var ErrEventParseError = errors.New("Could not parse event's payload")

type triggerFunc func(actions Actions)

func (f triggerFunc) Trigger(actions Actions) {
	f(actions)
}

// Event implements ActionsTrigger and has a sequence number
type Event struct {
	Seq int
	ActionsTrigger
}

// ParseEvent parses event's payload
func ParseEvent(payload []byte) (Event, error) {
	var (
		event  Event
		evType = []byte{0}
		arg1   int
		arg2   int
	)
	n, _ := fmt.Sscanf(string(payload), "%d|%1s|%d|%d\n", &event.Seq, &evType, &arg1, &arg2)
	type tuple struct {
		byte
		int
	}
	var trig triggerFunc
	switch (tuple{evType[0], n}) {
	case tuple{'F', 4}:
		trig = func(actions Actions) {
			actions.Follow(arg1, arg2)
			actions.SendMsg(arg2, payload)
		}
	case tuple{'U', 4}:
		trig = func(actions Actions) {
			actions.Unfollow(arg1, arg2)
		}
	case tuple{'B', 2}:
		trig = func(actions Actions) {
			actions.Broadcast(payload)
		}
	case tuple{'P', 4}:
		trig = func(actions Actions) {
			actions.SendMsg(arg2, payload)
		}
	case tuple{'S', 3}:
		trig = func(actions Actions) {
			actions.SendMsgToFollowers(arg1, payload)
		}
	default:
		return event, ErrEventParseError
	}
	event.ActionsTrigger = trig
	return event, nil
}
