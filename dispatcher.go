package main

import (
	"errors"
)

var (
	// ErrSeqTooSmall is returned by Dispatch method of Dispatcher
	// to indicate that event sequence number is too small.
	ErrSeqTooSmall = errors.New("dispatcher: event.Seq too small")

	// ErrSeqTooLarge is returned by Dispatch method of Dispatcher
	// to indicate that event sequence number is too large.
	ErrSeqTooLarge = errors.New("dispatcher: event.Seq too large")

	// ErrSeqDuplicate is returned by Dispatch method of Dispatcher
	// to indicate that there has been an event with the same sequence number before.
	ErrSeqDuplicate = errors.New("dispatcher: event.Seq duplicate found")
)

// Dispatcher orders events and triggers their actions once they are in order.
type Dispatcher struct {
	startIndex   int
	currentIndex int
	actions      Actions
	triggers     []ActionsTrigger
}

// Dispatch inserts event e into the right slot and triggers actions of the ordered slice.
func (d *Dispatcher) Dispatch(e Event) error {
	i := e.Seq - d.startIndex
	if i < d.currentIndex {
		return ErrSeqTooSmall
	}
	l := len(d.triggers)
	if i >= d.currentIndex+l {
		return ErrSeqTooLarge
	}

	if d.triggers[i%l] != nil {
		return ErrSeqDuplicate
	}
	d.triggers[i%l] = e
	for {
		t := d.triggers[d.currentIndex%l]
		if t == nil {
			break
		}
		t.Trigger(d.actions)
		d.triggers[d.currentIndex%l] = nil
		d.currentIndex++
	}
	return nil
}

// Reset resets dispatcher's internal state.
func (d *Dispatcher) Reset() {
	for i := range d.triggers {
		d.triggers[i] = nil
	}
	d.currentIndex = 0
}

// NewDispatcher returns a new Dispatcher.
func NewDispatcher(a Actions, startIndex, capacity int) *Dispatcher {
	return &Dispatcher{
		startIndex: startIndex,
		actions:    a,
		triggers:   make([]ActionsTrigger, capacity),
	}
}
