package reaction

import "testing"

func TestParseReactionNameCustom(t *testing.T) {
	got, err := ParseReactionName("party%3A2230469276416868352")
	if err != nil {
		t.Fatalf("ParseReactionName returned error: %v", err)
	}
	if !got.Custom {
		t.Fatalf("expected custom reaction, got %#v", got)
	}
	if got.EmojiId != 2230469276416868352 {
		t.Fatalf("unexpected emoji id: %d", got.EmojiId)
	}
	if got.EmojiName != "party" {
		t.Fatalf("unexpected emoji name: %q", got.EmojiName)
	}
	if got.BucketKey != "c:2230469276416868352" {
		t.Fatalf("unexpected bucket key: %q", got.BucketKey)
	}
}

func TestParseReactionNameUnicode(t *testing.T) {
	got, err := ParseReactionName("%E2%9D%A4%EF%B8%8F")
	if err != nil {
		t.Fatalf("ParseReactionName returned error: %v", err)
	}
	if got.Custom {
		t.Fatalf("expected unicode reaction, got %#v", got)
	}
	if got.EmojiName != "❤️" {
		t.Fatalf("unexpected emoji name: %q", got.EmojiName)
	}
	if got.BucketKey != "u:❤️" {
		t.Fatalf("unexpected bucket key: %q", got.BucketKey)
	}
}
