package router

import (
	"errors"
	"sync"

	"github.com/telendt/fmaze/event"
)

var (
	// ErrChannelAlreadySubscribed is returned by Router.Subscribe calls
	// when client's send channel has been already subscribed.
	ErrChannelAlreadySubscribed = errors.New("client already subscribed")

	rt *Router
	_  event.Actions = rt
)

// cSet represents set of send channels and provides some utility methods.
type cSet map[chan<- []byte]struct{}

func (s cSet) add(c chan<- []byte) {
	s[c] = struct{}{}
}

func (s cSet) extend(other cSet) {
	for c := range other {
		s.add(c)
	}
}

func (s cSet) remove(c chan<- []byte) {
	delete(s, c)
}

func (s cSet) subtract(other cSet) {
	for c := range other {
		s.remove(c)
	}
}

type cSetsMap map[int]cSet

func (m cSetsMap) getOrCreate(userID int) cSet {
	conns, ok := m[userID]
	if !ok {
		conns = make(cSet)
		m[userID] = conns
	}
	return conns
}

func (m cSetsMap) removeMember(userID int, c chan<- []byte) {
	if conns, ok := m[userID]; ok {
		conns.remove(c)
		if len(conns) == 0 {
			delete(m, userID)
		}
	}
}

func (m cSetsMap) removeMembers(userID int, cset cSet) {
	if conns, ok := m[userID]; ok {
		conns.subtract(cset)
		if len(conns) == 0 {
			delete(m, userID)
		}
	}
}

// Router implements Actions interface.
type Router struct {
	mu sync.RWMutex

	// closed on reset
	done chan struct{}

	sendToAll func([]byte, cSet)

	connectedClients   cSetsMap
	connectedFollowers cSetsMap
	allConnected       cSet

	// inverted connection graph
	invGraph Graph
}

// New returns new Router.
func New(blockingSend bool) *Router {
	f := func(msg []byte, s cSet) {
		for c := range s {
			select {
			case c <- msg:
			default:
			}
		}
	}
	if blockingSend {
		f = func(msg []byte, s cSet) {
			switch l := len(s); {
			case l == 1:
				for c := range s {
					c <- msg
				}
			case l > 1:
				var wg sync.WaitGroup
				for c := range s {
					wg.Add(1)
					go func(c chan<- []byte) {
						defer wg.Done()
						c <- msg
					}(c)
				}
				wg.Wait()
			default: // l == 0
			}
		}
	}

	return &Router{
		done:               make(chan struct{}),
		sendToAll:          f,
		connectedClients:   make(cSetsMap),
		connectedFollowers: make(cSetsMap),
		allConnected:       make(cSet),
		invGraph:           make(sparseGraph),
	}
}

// Reset resets inverted connections graph.
func (g *Router) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	close(g.done)
	g.done = make(chan struct{})
	g.invGraph = make(sparseGraph)
}

// Subscribe adds user client (its send channel) to Router and returns UnsubscribeFunc.
// It also returns ErrChannelAlreadySubscribed if the channel has already been subsribed to any userID.
// Given channel can only subscribe to a single userID, but it's fine to subscribe
// multiple different channels under the same userID.
func (g *Router) Subscribe(userID int, c chan<- []byte) (UnsubscribeFunc, <-chan struct{}, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.allConnected[c]; ok {
		return nil, g.done, ErrChannelAlreadySubscribed
	}
	g.connectedClients.getOrCreate(userID).add(c)
	for id := range g.invGraph.Neighbors(userID) {
		g.connectedFollowers.getOrCreate(id).add(c)
	}
	g.allConnected.add(c)

	ig := g.invGraph

	cleanup := func() {
		g.connectedClients.removeMember(userID, c)
		for id := range ig.Neighbors(userID) {
			g.connectedFollowers.removeMember(id, c)
		}
		delete(g.allConnected, c)
	}

	return func() {
		g.mu.Lock()
		defer g.mu.Unlock()

		if cleanup != nil {
			cleanup()
			cleanup = nil
		}
	}, g.done, nil
}

// Follow adds followerID to list of followers of user identified by followedID.
func (g *Router) Follow(followerID, followedID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.invGraph.Connect(followerID, followedID)
	if conns, ok := g.connectedClients[followerID]; ok {
		g.connectedFollowers.getOrCreate(followedID).extend(conns)
	}
}

// Unfollow removes followerID from list of followers of user identified by followedID.
func (g *Router) Unfollow(followerID, followedID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.invGraph.Disconnect(followerID, followedID)
	if conns, ok := g.connectedClients[followerID]; ok {
		g.connectedFollowers.removeMembers(followedID, conns)
	}
}

// SendMsg sends message msg to connected clients registered with userID identifier.
func (g *Router) SendMsg(userID int, msg []byte) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if conns, ok := g.connectedClients[userID]; ok {
		g.sendToAll(msg, conns)
	}
}

// SendMsgToFollowers sends message msg to connected followers of user identified by userID.
func (g *Router) SendMsgToFollowers(userID int, msg []byte) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if conns, ok := g.connectedFollowers[userID]; ok {
		g.sendToAll(msg, conns)
	}
}

// Broadcast sends message msg to all connected users.
func (g *Router) Broadcast(msg []byte) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	g.sendToAll(msg, g.allConnected)
}
