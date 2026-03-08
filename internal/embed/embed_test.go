package embed

import "testing"

func TestValidateEmbedsAcceptsDiscordVideoPayload(t *testing.T) {
	color := 16711680
	width := int64(1280)
	height := int64(720)
	placeholderVersion := 1
	flags := 0

	embeds := []Embed{{
		Type:        "video",
		URL:         "https://www.youtube.com/watch?v=OgfdyH4iaps",
		Title:       "Why is Microsoft updating their text editors!? | TheStandup",
		Description: "Chapters\n00:00:00 - Intro",
		Color:       &color,
		Author: &EmbedAuthor{
			Name: "The PrimeTime",
			URL:  "https://www.youtube.com/channel/UCUyeluBRhGPCW4rPe_UvBZQ",
		},
		Provider: &EmbedProvider{
			Name: "YouTube",
			URL:  "https://www.youtube.com",
		},
		Thumbnail: &EmbedMedia{
			URL:                "https://i.ytimg.com/vi/OgfdyH4iaps/maxresdefault.jpg",
			ProxyURL:           "https://images-ext-1.discordapp.net/external/jm6iBEsldkSfkcx2xcYp6-x-dShUdNSMiWSY3ejdkME/https/i.ytimg.com/vi/OgfdyH4iaps/maxresdefault.jpg",
			Width:              &width,
			Height:             &height,
			ContentType:        "image/jpeg",
			Placeholder:        "IPgNDIQbRUaQiJcwuYlwmyQ5Bg==",
			PlaceholderVersion: &placeholderVersion,
			Flags:              &flags,
		},
		Video: &EmbedMedia{
			URL:                "https://www.youtube.com/embed/OgfdyH4iaps",
			Width:              &width,
			Height:             &height,
			Placeholder:        "IPgNDIQbRUaQiJcwuYlwmyQ5Bg==",
			PlaceholderVersion: &placeholderVersion,
			Flags:              &flags,
		},
	}}

	if err := ValidateEmbeds(embeds); err != nil {
		t.Fatalf("ValidateEmbeds returned error: %v", err)
	}

	raw, err := MarshalEmbeds(embeds)
	if err != nil {
		t.Fatalf("MarshalEmbeds returned error: %v", err)
	}

	parsed, err := ParseEmbeds(&raw)
	if err != nil {
		t.Fatalf("ParseEmbeds returned error: %v", err)
	}
	if len(parsed) != 1 || parsed[0].Type != "video" {
		t.Fatalf("unexpected parsed embeds: %#v", parsed)
	}
}

func TestValidateEmbedsRejectsTooManyEmbeds(t *testing.T) {
	embeds := make([]Embed, MaxEmbedsPerMessage+1)
	for i := range embeds {
		embeds[i] = Embed{Description: "hello"}
	}

	if err := ValidateEmbeds(embeds); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateEmbedsRejectsNegativeFlags(t *testing.T) {
	flags := -1
	embeds := []Embed{{
		Image: &EmbedMedia{
			URL:   "https://example.com/image.png",
			Flags: &flags,
		},
	}}

	if err := ValidateEmbeds(embeds); err == nil {
		t.Fatal("expected validation error")
	}
}
