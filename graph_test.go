package main

import (
	"reflect"
	"testing"
)

func TestGraphSubscribeUnsubscribe(t *testing.T) {
	g := NewUserGraph()
	c := make(chan []byte)
	u, err := g.Subscribe(1, c)
	if err != nil {
		t.Errorf("First subscribe returned error %s", err.Error())
	}
	if len(g.allConnected) != 1 {
		t.Error("Subscribe should add connection to allConnected")
	}
	if _, err := g.Subscribe(2, c); err != ErrChannelAlreadySubscribed {
		t.Error("Second subscribe didn't return ErrChannelAlreadySubscribed error")
	}
	if len(g.allConnected) != 1 {
		t.Error("allConnected should not change")
	}
	u()
	if len(g.allConnected) != 0 {
		t.Error("allConnected should be empty")
	}
	u() // should be a NOOP at this point
}

func TestGraphActions(t *testing.T) {
	g := NewUserGraph()

	c1 := make(chan []byte, 1)
	c2 := make(chan []byte, 1)
	c3 := make(chan []byte, 1)
	g.Subscribe(1, c1)
	g.Subscribe(2, c2)
	g.Subscribe(3, c3)

	msg := []byte("msg")
	receivedMsg := func(c <-chan []byte) bool {
		select {
		case m := <-c:
			if !reflect.DeepEqual(m, msg) {
				t.Fatalf("Received incorrect message, %#v != %#v", m, msg)
			}
			return true
		default:
			return false
		}
	}

	g.SendMsg(2, msg)
	if a, b, c := receivedMsg(c1), receivedMsg(c2), receivedMsg(c3); a || !b || c {
		t.Errorf("Only client 2 should receive a message (%v, %v, %v)", a, b, c)
	}

	g.Broadcast(msg)
	if a, b, c := receivedMsg(c1), receivedMsg(c2), receivedMsg(c3); !a || !b || !c {
		t.Errorf("All clients should receive a message (%v, %v, %v)", a, b, c)
	}

	g.Follow(1, 3)
	g.Follow(2, 3)
	g.Follow(3, 2)
	g.SendMsgToFollowers(3, msg)
	if a, b, c := receivedMsg(c1), receivedMsg(c2), receivedMsg(c3); !a || !b || c {
		t.Errorf("Only clients 1 and 2 (followers of 3) should receive a message (%v, %v, %v)", a, b, c)
	}

	g.Unfollow(2, 3)
	g.SendMsgToFollowers(3, msg)
	if a, b, c := receivedMsg(c1), receivedMsg(c2), receivedMsg(c3); !a || b || c {
		t.Errorf("Only client 1 (follower of 3) should receive a message (%v, %v, %v)", a, b, c)
	}

	// NOOPS
	g.SendMsg(4, msg)
	g.Unfollow(2, 3)

	// first follow, then subscribe
	g.Follow(4, 1)
	c4 := make(chan []byte, 1)
	g.Subscribe(4, c4)
	g.SendMsgToFollowers(1, msg)
	if a, b, c, d := receivedMsg(c1), receivedMsg(c2), receivedMsg(c3), receivedMsg(c4); a || b || c || !d {
		t.Errorf("Only client 4 (follower of 1) should receive a message (%v, %v, %v, %v)", a, b, c, d)
	}
}
