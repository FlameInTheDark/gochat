package emoji

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

const (
	MaxUploadSizeBytes      int64 = 256 * 1024
	MaxDimension                  = 128
	MaxStaticPerGuild             = 50
	MaxAnimatedPerGuild           = 50
	MaxActivePerGuild             = 100
	LookupCacheTTLSeconds   int64 = 3600
	NegativeCacheTTLSeconds int64 = 60
	GuildCacheTTLSeconds    int64 = 600
)

var NameRegex = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

type LookupCacheEntry struct {
	Missing  bool   `json:"missing,omitempty"`
	Id       int64  `json:"id"`
	GuildId  int64  `json:"guild_id"`
	Name     string `json:"name"`
	Done     bool   `json:"done"`
	Animated bool   `json:"animated"`
	Width    int64  `json:"width,omitempty"`
	Height   int64  `json:"height,omitempty"`
}

func NormalizeName(name string) string {
	return strings.ToLower(name)
}

func LookupCacheKey(emojiID int64) string {
	return fmt.Sprintf("emoji:id:%d", emojiID)
}

func GuildCacheKey(guildID int64) string {
	return fmt.Sprintf("emoji:guild:%d", guildID)
}

func SelectClosestVariant(size int) string {
	if size <= 0 {
		return "master"
	}

	candidates := []struct {
		name  string
		size  int
		score int
	}{
		{name: "44", size: 44},
		{name: "96", size: 96},
		{name: "master", size: 128},
	}

	best := candidates[0]
	best.score = int(math.Abs(float64(size - best.size)))
	for i := 1; i < len(candidates); i++ {
		candidate := candidates[i]
		candidate.score = int(math.Abs(float64(size - candidate.size)))
		if candidate.score < best.score || (candidate.score == best.score && candidate.size > best.size) {
			best = candidate
		}
	}
	return best.name
}
