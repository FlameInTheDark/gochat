package botgateway

import "testing"

func TestPartitionForIDStable(t *testing.T) {
	first := PartitionForID(42, 1024)
	second := PartitionForID(42, 1024)
	if first != second {
		t.Fatalf("expected stable partition, got %d and %d", first, second)
	}
	if first < 0 || first >= 1024 {
		t.Fatalf("partition out of range: %d", first)
	}
}

func TestParseEventSubject(t *testing.T) {
	got, ok := ParseEventSubject(GuildChannelSubject(42, 99, 1024))
	if !ok || got.Kind != EventKindGuildChannel || got.GuildID != 42 || got.ChannelID != 99 {
		t.Fatalf("unexpected guild channel subject parse: %#v ok=%v", got, ok)
	}

	got, ok = ParseEventSubject(UserDMSubject(77, 1024))
	if !ok || got.Kind != EventKindUserDM || got.UserID != 77 {
		t.Fatalf("unexpected user dm subject parse: %#v ok=%v", got, ok)
	}
}

func TestTargetSessions(t *testing.T) {
	sessions := []Session{
		{SessionID: "unsharded-a", ShardCount: 1, ShardID: 0, ReceivesDMEvents: true},
		{SessionID: "unsharded-b", ShardCount: 1, ShardID: 0, ReceivesDMEvents: true},
		{SessionID: "shard-0", ShardCount: 4, ShardID: 0, ReceivesDMEvents: true},
		{SessionID: "shard-2", ShardCount: 4, ShardID: 2, ReceivesDMEvents: false},
	}

	guildTargets := TargetGuildSessions(sessions, 10)
	if len(guildTargets) != 3 {
		t.Fatalf("expected two unsharded sessions plus shard 2, got %#v", guildTargets)
	}

	dmTargets := TargetDMSessions(sessions)
	if len(dmTargets) != 3 {
		t.Fatalf("expected two unsharded sessions plus shard 0, got %#v", dmTargets)
	}
}
