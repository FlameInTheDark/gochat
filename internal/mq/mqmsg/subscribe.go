package mqmsg

type Subscribe struct {
	Channel *int64  `json:"channel,omitempty"`
	Guilds  []int64 `json:"guilds,omitempty"`
}
