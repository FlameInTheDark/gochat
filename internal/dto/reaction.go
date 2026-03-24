package dto

type MessageReaction struct {
	Count int                  `json:"count"`
	Me    bool                 `json:"me"`
	Emoji MessageReactionEmoji `json:"emoji"`
}

type MessageReactionEmoji struct {
	Id   *int64 `json:"id"`
	Name string `json:"name"`
}

type MessageReactionUsersPage struct {
	Items     []User `json:"items"`
	NextAfter *int64 `json:"next_after,omitempty"`
}
