package router

// UnsubscribeFunc unsubscribes previously subscribed channel connection.
type UnsubscribeFunc func()

// Subscriber is the interface implemented by UserGraph that wraps the basic Subscribe method.
//
// Subscribe subscribes given channel c under identifier id and returns unsubscribe
// function (that takes no arguments), empty struct channel (used to broadcast a done signal)
// and any error that prevented successful subscription.
type Subscriber interface {
	Subscribe(id int, c chan<- []byte) (UnsubscribeFunc, <-chan struct{}, error)
}
