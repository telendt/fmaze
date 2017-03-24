package main

// Actions define possible event actions
type Actions interface {
	// graph mutating actions
	Follow(followerID, followedID int)
	Unfollow(followerID, followedID int)

	// notify actions
	SendMsg(userID int, msg []byte)
	SendMsgToFollowers(userID int, msg []byte)
	Broadcast(msg []byte)
}

// ActionsTrigger interface defines single Trigger method, that triggers one or more actions.
type ActionsTrigger interface {
	Trigger(Actions)
}

// UnsubscribeFunc unsubscribes previously subscribed channel connection.
type UnsubscribeFunc func()

// Subscriber is the interface implemented by UserGraph that wraps the basic Subscribe method.
//
// Subscribe subscribes given channel c under identifier id and returns unsubscibe
// function (that takes no arguments) and any error that prevented successful subscription.
type Subscriber interface {
	Subscribe(id int, c chan<- []byte) (UnsubscribeFunc, error)
}
