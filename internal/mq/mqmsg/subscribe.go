package mqmsg

type Subscribe struct {
	Channel  *int64  `json:"channel,omitempty"`
	Channels []int64 `json:"channels,omitempty"`
	Guilds   []int64 `json:"guilds,omitempty"`
}
