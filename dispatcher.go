package main

// Dispatcher dispatchers events
type Dispatcher struct {
	startIndex int
	i          int
	actions    Actions
	triggers   []ActionsTrigger
}

func (d *Dispatcher) Dispatch(e Event) {
	// TODO: handle seq > d.i + len(d.events)
	if e.Seq-d.startIndex < d.i {
		panic("seq too small")
	}
	l := len(d.triggers)
	d.triggers[(e.Seq-1)%l] = e
	for i := 0; i < l; i++ {
		t := d.triggers[d.i%l]
		if t == nil {
			break
		}
		t.Trigger(d.actions)
		d.triggers[d.i%l] = nil
		d.i++
	}
}

func (d *Dispatcher) Reset() {
	for i := range d.triggers {
		d.triggers[i] = nil
	}
	d.i = 0
}

func NewDispatcher(a Actions, startIndex, capacity int) *Dispatcher {
	return &Dispatcher{
		startIndex: startIndex,
		actions:    a,
		triggers:   make([]ActionsTrigger, capacity),
	}
}
