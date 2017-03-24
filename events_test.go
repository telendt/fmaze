package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type actionCall string

func followCall(a1, a2 int) actionCall {
	return actionCall(fmt.Sprintf("Follow(%#v, %#v)", a1, a2))
}

func unfollowCall(a1, a2 int) actionCall {
	return actionCall(fmt.Sprintf("Unfollow(%#v, %#v)", a1, a2))
}

func sendMsgCall(a1 int, a2 []byte) actionCall {
	return actionCall(fmt.Sprintf("SendMsg(%#v, %#v)", a1, a2))
}

func sendMsgToFollowersCall(a1 int, a2 []byte) actionCall {
	return actionCall(fmt.Sprintf("SendMsgToFollowers(%#v, %#v)", a1, a2))
}

func broadcastCall(a1 []byte) actionCall {
	return actionCall(fmt.Sprintf("Broadcast(%#v)", a1))
}

type actionsCallSpy struct {
	callStack []actionCall
}

func (a *actionsCallSpy) Follow(a1, a2 int) {
	a.callStack = append(a.callStack, followCall(a1, a2))
}

func (a *actionsCallSpy) Unfollow(a1, a2 int) {
	a.callStack = append(a.callStack, unfollowCall(a1, a2))
}

func (a *actionsCallSpy) SendMsg(a1 int, a2 []byte) {
	a.callStack = append(a.callStack, sendMsgCall(a1, a2))
}

func (a *actionsCallSpy) SendMsgToFollowers(a1 int, a2 []byte) {
	a.callStack = append(a.callStack, sendMsgToFollowersCall(a1, a2))
}

func (a *actionsCallSpy) Broadcast(a1 []byte) {
	a.callStack = append(a.callStack, broadcastCall(a1))
}

func fmtCalls(actionCalls []actionCall) string {
	calls := make([]string, len(actionCalls))
	for i, c := range actionCalls {
		calls[i] = string(c)
	}
	return fmt.Sprintf("{%s}", strings.Join(calls, ", "))
}

func TestParseEventsSuccess(t *testing.T) {
	for _, testCase := range []struct {
		payloadStr string
		seq        int
		calls      []actionCall
	}{
		{"666|F|60|50\n", 666, []actionCall{
			followCall(60, 50),
			sendMsgCall(50, []byte("666|F|60|50\n")),
		}},
		{"1|U|12|9\n", 1, []actionCall{
			unfollowCall(12, 9),
		}},
		{"542532|B\n", 542532, []actionCall{
			broadcastCall([]byte("542532|B\n")),
		}},
		{"43|P|32|56\n", 43, []actionCall{
			sendMsgCall(56, []byte("43|P|32|56\n")),
		}},
		{"634|S|32\n", 634, []actionCall{
			sendMsgToFollowersCall(32, []byte("634|S|32\n")),
		}},
	} {
		event, err := ParseEvent([]byte(testCase.payloadStr))
		if err != nil {
			t.Errorf("%s: %s", testCase.payloadStr, err.Error())
			continue
		}
		if event.Seq != testCase.seq {
			t.Errorf("%s: want sequence: %d, have: %d", testCase.payloadStr, testCase.seq, event.Seq)
			continue
		}
		spy := &actionsCallSpy{}
		event.Trigger(spy)
		if !reflect.DeepEqual(spy.callStack, testCase.calls) {
			t.Errorf("%s: calls %s != %s", testCase.payloadStr, fmtCalls(spy.callStack), fmtCalls(testCase.calls))
		}
	}

}
