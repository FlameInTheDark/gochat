package embedgen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateFromOpenGraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<!doctype html><html><head>
			<meta property="og:title" content="Example title">
			<meta property="og:description" content="Example description">
			<meta property="og:site_name" content="Example Site">
			<meta property="og:image" content="/preview.png">
			<meta property="og:image:width" content="1200">
			<meta property="og:image:height" content="630">
		</head></html>`)
	}))
	defer server.Close()

	generator := New(Config{AllowPrivateHosts: true})
	embeds, err := generator.Generate(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embeds))
	}
	if embeds[0].Title != "Example title" {
		t.Fatalf("unexpected title: %q", embeds[0].Title)
	}
	if embeds[0].Provider == nil || embeds[0].Provider.Name != "Example Site" {
		t.Fatalf("unexpected provider: %#v", embeds[0].Provider)
	}
	if embeds[0].Thumbnail == nil || embeds[0].Thumbnail.URL == "" {
		t.Fatalf("unexpected thumbnail: %#v", embeds[0].Thumbnail)
	}
	if embeds[0].Type != "link" {
		t.Fatalf("unexpected type: %q", embeds[0].Type)
	}
}

func TestGenerateFromOEmbedDiscovery(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":             "rich",
				"title":            "Rich title",
				"author_name":      "Author",
				"provider_name":    "Provider",
				"provider_url":     server.URL,
				"thumbnail_url":    server.URL + "/thumb.png",
				"thumbnail_width":  640,
				"thumbnail_height": 360,
			})
		default:
			fmt.Fprintf(w, `<!doctype html><html><head>
				<link rel="alternate" type="application/json+oembed" href="%s/oembed">
				<meta property="og:description" content="Fallback description">
			</head></html>`, server.URL)
		}
	}))
	defer server.Close()

	generator := New(Config{AllowPrivateHosts: true})
	embeds, err := generator.Generate(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embeds))
	}
	if embeds[0].Title != "Rich title" {
		t.Fatalf("unexpected title: %q", embeds[0].Title)
	}
	if embeds[0].Author == nil || embeds[0].Author.Name != "Author" {
		t.Fatalf("unexpected author: %#v", embeds[0].Author)
	}
	if embeds[0].Provider == nil || embeds[0].Provider.Name != "Provider" {
		t.Fatalf("unexpected provider: %#v", embeds[0].Provider)
	}
}

func TestGenerateFromYouTubeURL(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtube/oembed" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("url"); got == "" {
			t.Fatal("expected oembed url query")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":             "video",
			"title":            "Video title",
			"author_name":      "Channel",
			"author_url":       server.URL + "/channel",
			"provider_name":    "YouTube",
			"provider_url":     "https://www.youtube.com",
			"thumbnail_url":    server.URL + "/thumb.jpg",
			"thumbnail_width":  1280,
			"thumbnail_height": 720,
			"width":            1280,
			"height":           720,
		})
	}))
	defer server.Close()

	generator := New(Config{
		AllowPrivateHosts:     true,
		YouTubeOEmbedEndpoint: server.URL + "/youtube/oembed",
	})
	generated, err := generator.GenerateURL(context.Background(), "https://www.youtube.com/watch?v=abc123xyz")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.Type != "video" {
		t.Fatalf("unexpected type: %q", generated.Type)
	}
	if generated.Video == nil || generated.Video.URL != "https://www.youtube.com/embed/abc123xyz" {
		t.Fatalf("unexpected video payload: %#v", generated.Video)
	}
	if generated.Provider == nil || generated.Provider.Name != "YouTube" {
		t.Fatalf("unexpected provider: %#v", generated.Provider)
	}
	if generated.Author == nil || generated.Author.Name != "Channel" {
		t.Fatalf("unexpected author: %#v", generated.Author)
	}
	if generated.Color == nil || *generated.Color != 0xFF0000 {
		t.Fatalf("unexpected youtube color: %#v", generated.Color)
	}
}

func TestGenerateFromYouTubeURLWithCustomEmbedBase(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":             "video",
			"title":            "Video title",
			"provider_name":    "YouTube",
			"provider_url":     "https://www.youtube.com",
			"thumbnail_url":    server.URL + "/thumb.jpg",
			"thumbnail_width":  1280,
			"thumbnail_height": 720,
			"width":            1280,
			"height":           720,
		})
	}))
	defer server.Close()

	generator := New(Config{
		AllowPrivateHosts:     true,
		YouTubeOEmbedEndpoint: server.URL,
		YouTubeEmbedBaseURL:   "https://www.youtube-nocookie.com/embed",
	})
	generated, err := generator.GenerateURL(context.Background(), "https://www.youtube.com/watch?v=abc123xyz")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil || generated.Video == nil {
		t.Fatalf("expected generated video embed, got %#v", generated)
	}
	if generated.Video.URL != "https://www.youtube-nocookie.com/embed/abc123xyz" {
		t.Fatalf("unexpected custom video url: %q", generated.Video.URL)
	}
}

func TestGenerateTwitterEmbedUsesBrandColor(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":             "rich",
				"title":            "Tweet title",
				"provider_name":    "Twitter",
				"provider_url":     "https://twitter.com",
				"author_name":      "example",
				"thumbnail_url":    server.URL + "/thumb.png",
				"thumbnail_width":  640,
				"thumbnail_height": 360,
			})
		default:
			fmt.Fprintf(w, `<!doctype html><html><head>
				<link rel="alternate" type="application/json+oembed" href="%s/oembed">
			</head></html>`, server.URL)
		}
	}))
	defer server.Close()

	generator := New(Config{AllowPrivateHosts: true})
	generated, err := generator.GenerateURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.Color == nil || *generated.Color != 0x1DA1F2 {
		t.Fatalf("unexpected twitter color: %#v", generated.Color)
	}
}
