package goclient

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	// APIVersion is the GoChat bot REST API version used by this client.
	APIVersion = "v1"
	// EndpointAPIPath is the base path for bot REST requests.
	EndpointAPIPath = "/bot/api/" + APIVersion
	// EndpointGatewayPath is the bot gateway WebSocket path.
	EndpointGatewayPath = "/bot/ws"
)

// Known GoChat bot API paths.
var (
	EndpointUserMe = EndpointAPIPath + "/user/me"

	EndpointGuilds        = EndpointAPIPath + "/guild"
	EndpointGuildChannels = func(guildID string) string { return EndpointGuilds + "/" + guildID + "/channels" }

	EndpointChannelMessages   = func(channelID string) string { return EndpointAPIPath + "/message/channel/" + channelID }
	EndpointChannelMessage    = func(channelID, messageID string) string { return EndpointChannelMessages(channelID) + "/" + messageID }
	EndpointChannelMessageAck = func(channelID, messageID string) string { return EndpointChannelMessage(channelID, messageID) + "/ack" }
	EndpointChannelTyping     = func(channelID string) string { return EndpointChannelMessages(channelID) + "/typing" }
	EndpointMessageReactions  = func(channelID, messageID, reaction string) string {
		return EndpointChannelMessage(channelID, messageID) + "/reactions/" + url.PathEscape(reaction)
	}
	EndpointMessageReactionAll = func(channelID, messageID string) string {
		return EndpointChannelMessage(channelID, messageID) + "/reactions"
	}
)

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return DefaultEndpoint
	}
	return strings.TrimRight(endpoint, "/")
}

func gatewayEndpointFromBase(endpoint string) string {
	endpoint = normalizeEndpoint(endpoint)
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimRight(endpoint, "/") + EndpointGatewayPath
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	u.Path = strings.TrimRight(u.Path, "/") + EndpointGatewayPath
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
