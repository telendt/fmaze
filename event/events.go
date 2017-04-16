package event

import (
	"errors"
	"fmt"
)

type event struct {
	eType        byte
	expectedArgs int
}

var (
	follow       = event{'F', 2}
	unfollow     = event{'U', 2}
	broadcast    = event{'B', 0}
	privateMsg   = event{'P', 2}
	statusUpdate = event{'S', 1}

	payloadFormat = "%d|%1s|%d|%d\n"
)

// ErrBadFormat is returned by Parse function when event's payload
// does not conform the expected format (`Seq|Type[|Arg1[|Arg2]]\n`).
var ErrBadFormat = errors.New("events: bad event payload format")

// UnknownTypeError records unknown event type.
type UnknownTypeError struct {
	Type byte
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("events: unknown type %q", e.Type)
}

// BadArgumentsNumberError records bad number of received arguments.
type BadArgumentsNumberError struct {
	Want int
	Got  int
}

func (e *BadArgumentsNumberError) Error() string {
	return fmt.Sprintf("events: expected %d arguments, got %d", e.Want, e.Got)
}

// Event implements ActionsTrigger and has a sequence number.
type Event struct {
	Seq int64
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

// Parse parses event's payload
func Parse(payload []byte) (Event, error) {
	var (
		e     Event
		eType = []byte{0}
		arg1  int
		arg2  int
	)
	n, _ := fmt.Sscanf(string(payload), payloadFormat, &e.Seq, &eType, &arg1, &arg2)
	if n < 2 {
		return e, ErrBadFormat
	}
	t := eType[0]
	nArgs := n - 2
	var trig ActionsTrigger
	switch t {
	case follow.eType:
		if nArgs != follow.expectedArgs {
			return e, &BadArgumentsNumberError{Want: follow.expectedArgs, Got: nArgs}
		}
		trig = followActionsTrigger{
			followerID: arg1,
			followedID: arg2,
			msg:        payload,
		}
	case unfollow.eType:
		if nArgs != unfollow.expectedArgs {
			return e, &BadArgumentsNumberError{Want: unfollow.expectedArgs, Got: nArgs}
		}
		trig = unfollowActionsTrigger{
			followerID: arg1,
			followedID: arg2,
		}
	case broadcast.eType:
		if nArgs != broadcast.expectedArgs {
			return e, &BadArgumentsNumberError{Want: broadcast.expectedArgs, Got: nArgs}
		}
		trig = broadcastActionsTrigger{
			msg: payload,
		}
	case privateMsg.eType:
		if nArgs != privateMsg.expectedArgs {
			return e, &BadArgumentsNumberError{Want: privateMsg.expectedArgs, Got: nArgs}
		}
		trig = privateMsgActionsTrigger{
			userID: arg2,
			msg:    payload,
		}
	case statusUpdate.eType:
		if nArgs != statusUpdate.expectedArgs {
			return e, &BadArgumentsNumberError{Want: statusUpdate.expectedArgs, Got: nArgs}
		}
		trig = statusUpdateActionsTrigger{
			userID: arg1,
			msg:    payload,
		}
	default:
		return e, &UnknownTypeError{eType[0]}
	}
	e.ActionsTrigger = trig
	return e, nil
}
