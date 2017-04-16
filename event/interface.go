package event

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
