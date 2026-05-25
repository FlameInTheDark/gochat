package userbootstrap

import (
	"sort"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

// FilterGuildLastMessages keeps only live non-thread channels from the guild
// last-message snapshot.
func FilterGuildLastMessages(glms map[int64]map[int64]int64, channels []model.Channel) map[int64]map[int64]int64 {
	if len(glms) == 0 || len(channels) == 0 {
		return map[int64]map[int64]int64{}
	}

	allowedChannels := make(map[int64]struct{}, len(channels))
	for _, channel := range channels {
		if channel.Type == model.ChannelTypeThread {
			continue
		}
		allowedChannels[channel.Id] = struct{}{}
	}

	filtered := make(map[int64]map[int64]int64, len(glms))
	for guildID, channelMessages := range glms {
		for channelID, lastMessageID := range channelMessages {
			if _, ok := allowedChannels[channelID]; !ok {
				continue
			}
			if lastMessageID <= channelID {
				continue
			}
			if filtered[guildID] == nil {
				filtered[guildID] = make(map[int64]int64)
			}
			filtered[guildID][channelID] = lastMessageID
		}
	}

	return filtered
}

// FilterThreadLastMessages keeps only live joined threads from the guild
// last-message snapshot. Threads without a parent are ignored so this map stays
// aligned with BuildJoinedThreads.
func FilterThreadLastMessages(joined map[int64]struct{}, channels []model.Channel, glms map[int64]map[int64]int64) map[int64]int64 {
	if len(joined) == 0 || len(channels) == 0 || len(glms) == 0 {
		return map[int64]int64{}
	}

	liveThreads := make(map[int64]struct{}, len(joined))
	for _, channel := range channels {
		if channel.Type != model.ChannelTypeThread || channel.ParentID == nil {
			continue
		}
		if _, ok := joined[channel.Id]; ok {
			liveThreads[channel.Id] = struct{}{}
		}
	}

	out := make(map[int64]int64, len(liveThreads))
	for _, channelMessages := range glms {
		for channelID, lastMessageID := range channelMessages {
			if _, ok := liveThreads[channelID]; !ok {
				continue
			}
			if lastMessageID <= channelID {
				continue
			}
			out[channelID] = lastMessageID
		}
	}
	return out
}

// BuildJoinedThreadSet deduplicates thread membership rows into a thread-id set.
func BuildJoinedThreadSet(threadMembers []model.ThreadMember) map[int64]struct{} {
	joined := make(map[int64]struct{}, len(threadMembers))
	for _, member := range threadMembers {
		if member.ThreadId <= 0 {
			continue
		}
		joined[member.ThreadId] = struct{}{}
	}
	return joined
}

// BuildJoinedThreads groups joined thread IDs as guild_id -> parent_channel_id
// -> sorted thread IDs.
func BuildJoinedThreads(joined map[int64]struct{}, channels []model.Channel, guildChannels []model.GuildChannel) map[int64]map[int64][]int64 {
	if len(joined) == 0 || len(channels) == 0 || len(guildChannels) == 0 {
		return map[int64]map[int64][]int64{}
	}

	guildByChannel := make(map[int64]int64, len(guildChannels))
	for _, guildChannel := range guildChannels {
		guildByChannel[guildChannel.ChannelId] = guildChannel.GuildId
	}

	out := make(map[int64]map[int64][]int64)
	for _, channel := range channels {
		if channel.Type != model.ChannelTypeThread || channel.ParentID == nil {
			continue
		}
		if _, ok := joined[channel.Id]; !ok {
			continue
		}
		guildID, ok := guildByChannel[channel.Id]
		if !ok {
			continue
		}
		if out[guildID] == nil {
			out[guildID] = make(map[int64][]int64)
		}
		parentID := *channel.ParentID
		out[guildID][parentID] = append(out[guildID][parentID], channel.Id)
	}

	for guildID, channelsMap := range out {
		for parentID, threads := range channelsMap {
			sort.Slice(threads, func(i, j int) bool {
				return threads[i] < threads[j]
			})
			channelsMap[parentID] = threads
		}
		out[guildID] = channelsMap
	}

	return out
}
