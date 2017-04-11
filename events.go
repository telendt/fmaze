package main

import (
	"errors"
	"fmt"
)

const (
	followEventType       = 'F'
	unfollowEventType     = 'U'
	broadcastEventType    = 'B'
	privateMsgEventType   = 'P'
	statusUpdateEventType = 'S'
)

// ErrEventParseError is returned by ParseEvent to indicate that problems with
// parsing event's payload.
var ErrEventParseError = errors.New("Could not parse event's payload")

// Event implements ActionsTrigger and has a sequence number
type Event struct {
	Seq int
	ActionsTrigger
}

type followActionsTrigger struct {
	followerID, followedID int
	msg                    []byte
}

func (f followActionsTrigger) Trigger(actions Actions) {
	actions.Follow(f.followerID, f.followedID)
	actions.SendMsg(f.followedID, f.msg)
}

type unfollowActionsTrigger struct {
	followerID, followedID int
}

func (u unfollowActionsTrigger) Trigger(actions Actions) {
	actions.Unfollow(u.followerID, u.followedID)
}

type broadcastActionsTrigger struct {
	msg []byte
}

func (b broadcastActionsTrigger) Trigger(actions Actions) {
	actions.Broadcast(b.msg)
}

type privateMsgActionsTrigger struct {
	userID int
	msg    []byte
}

func (p privateMsgActionsTrigger) Trigger(actions Actions) {
	actions.SendMsg(p.userID, p.msg)
}

type statusUpdateActionsTrigger struct {
	userID int
	msg    []byte
}

func (s statusUpdateActionsTrigger) Trigger(actions Actions) {
	actions.SendMsgToFollowers(s.userID, s.msg)
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
	var trig ActionsTrigger
	switch (tuple{evType[0], n}) {
	case tuple{followEventType, 4}:
		trig = followActionsTrigger{
			followerID: arg1,
			followedID: arg2,
			msg:        payload,
		}
	case tuple{unfollowEventType, 4}:
		trig = unfollowActionsTrigger{
			followerID: arg1,
			followedID: arg2,
		}
	case tuple{broadcastEventType, 2}:
		trig = broadcastActionsTrigger{
			msg: payload,
		}
	case tuple{privateMsgEventType, 4}:
		trig = privateMsgActionsTrigger{
			userID: arg2,
			msg:    payload,
		}
	case tuple{statusUpdateEventType, 3}:
		trig = statusUpdateActionsTrigger{
			userID: arg1,
			msg:    payload,
		}
	default:
		return event, ErrEventParseError
	}
	event.ActionsTrigger = trig
	return event, nil
}
