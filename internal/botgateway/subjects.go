package botgateway

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

const DefaultPartitions = 1024

type EventKind string

const (
	EventKindGuild        EventKind = "guild"
	EventKindGuildChannel EventKind = "guild_channel"
	EventKindUserDM       EventKind = "user_dm"
)

type EventSubject struct {
	Partition int
	Kind      EventKind
	GuildID   int64
	ChannelID int64
	UserID    int64
}

func NormalizePartitions(partitions int) int {
	if partitions <= 0 {
		return DefaultPartitions
	}
	return partitions
}

func PartitionForID(id int64, partitions int) int {
	partitions = NormalizePartitions(partitions)
	h := fnv.New64a()
	_, _ = h.Write([]byte(strconv.FormatInt(id, 10)))
	return int(h.Sum64() % uint64(partitions))
}

func GuildSubject(guildID int64, partitions int) string {
	partition := PartitionForID(guildID, partitions)
	return fmt.Sprintf("bot.event.p.%d.guild.%d", partition, guildID)
}

func GuildChannelSubject(guildID, channelID int64, partitions int) string {
	partition := PartitionForID(guildID, partitions)
	return fmt.Sprintf("bot.event.p.%d.guild.%d.channel.%d", partition, guildID, channelID)
}

func UserDMSubject(botUserID int64, partitions int) string {
	partition := PartitionForID(botUserID, partitions)
	return fmt.Sprintf("bot.event.p.%d.user.%d.dm", partition, botUserID)
}

func PartitionWildcardSubject(partition int) string {
	return fmt.Sprintf("bot.event.p.%d.>", partition)
}

func InstanceDeliverySubject(instanceID string) string {
	return "bot.deliver.instance." + instanceID
}

func ParseEventSubject(subject string) (EventSubject, bool) {
	parts := strings.Split(subject, ".")
	if len(parts) < 5 || parts[0] != "bot" || parts[1] != "event" || parts[2] != "p" {
		return EventSubject{}, false
	}
	partition, err := strconv.Atoi(parts[3])
	if err != nil || partition < 0 {
		return EventSubject{}, false
	}
	switch parts[4] {
	case "guild":
		if len(parts) != 6 && len(parts) != 8 {
			return EventSubject{}, false
		}
		guildID, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil || guildID <= 0 {
			return EventSubject{}, false
		}
		if len(parts) == 6 {
			return EventSubject{Partition: partition, Kind: EventKindGuild, GuildID: guildID}, true
		}
		if parts[6] != "channel" {
			return EventSubject{}, false
		}
		channelID, err := strconv.ParseInt(parts[7], 10, 64)
		if err != nil || channelID <= 0 {
			return EventSubject{}, false
		}
		return EventSubject{Partition: partition, Kind: EventKindGuildChannel, GuildID: guildID, ChannelID: channelID}, true
	case "user":
		if len(parts) != 7 || parts[6] != "dm" {
			return EventSubject{}, false
		}
		userID, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil || userID <= 0 {
			return EventSubject{}, false
		}
		return EventSubject{Partition: partition, Kind: EventKindUserDM, UserID: userID}, true
	default:
		return EventSubject{}, false
	}
}
