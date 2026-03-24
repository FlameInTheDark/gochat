package model

type Reaction struct {
	MessageId  int64
	ReactionId int64
	UserId     int64
	BucketKey  string
	Custom     bool
	EmojiId    int64
	EmojiName  string
}

type ReactionSummary struct {
	MessageId int64
	BucketKey string
	Custom    bool
	EmojiId   int64
	EmojiName string
	Count     int
}
