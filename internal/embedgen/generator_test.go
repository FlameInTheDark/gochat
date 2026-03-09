package embedgen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func mustNewGenerator(t *testing.T, cfg Config) *Generator {
	t.Helper()

	generator, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return generator
}

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

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})
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

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})
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

	generator := mustNewGenerator(t, Config{
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

	generator := mustNewGenerator(t, Config{
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

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})
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

func TestDescriptionFromOEmbedHTML(t *testing.T) {
	description := descriptionFromOEmbedHTML(`<blockquote class="twitter-tweet"><p>Tweet body <a href="https://t.co/test">https://t.co/test</a></p></blockquote>`)
	if description != "Tweet body https://t.co/test" {
		t.Fatalf("unexpected description: %q", description)
	}
}

func TestProxyTwitterVideoURL(t *testing.T) {
	proxied := proxyTwitterVideoURL(
		"https://vxtwitter.com",
		"https://video.twimg.com/ext_tw_video/2029212410624552960/pu/vid/avc1/720x1280/waHcL92w7O7whM7F.mp4?tag=12",
	)
	if proxied != "https://vxtwitter.com/tvid/ext_tw_video/2029212410624552960/pu/vid/avc1/720x1280/waHcL92w7O7whM7F" {
		t.Fatalf("unexpected proxied url: %q", proxied)
	}
}

func TestGenerateVXTwitterEmbedUsesServiceAPI(t *testing.T) {
	generator := mustNewGenerator(t, Config{
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Host {
				case "api.fxtwitter.com":
					if r.URL.Path != "/example/status/1234567890" {
						t.Fatalf("unexpected api path: %s", r.URL.Path)
					}
					body := `{
						"code": 200,
						"tweet": {
							"url": "https://x.com/example/status/1234567890",
							"id": "1234567890",
							"text": "Proxy tweet text",
							"created_timestamp": 1700000000,
							"author": {
								"name": "Example Display",
								"screen_name": "example",
								"avatar_url": "https://pbs.twimg.com/profile_images/example_normal.jpg"
							},
							"media": {
								"videos": [{
									"type": "video",
									"url": "https://video.twimg.com/ext_tw_video/1234567890/pu/vid/avc1/720x1280/video.mp4",
									"thumbnail_url": "https://pbs.twimg.com/ext_tw_video_thumb/1234567890/pu/img/thumb.jpg",
									"width": 720,
									"height": 1280,
									"format": "video/mp4"
								}]
							}
						}
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
						Body:       io.NopCloser(bytes.NewBufferString(body)),
						Request:    r,
					}, nil
				case "pbs.twimg.com":
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"image/jpeg"}},
						Body:       io.NopCloser(bytes.NewReader(pngPayload(675, 1200))),
						Request:    r,
					}, nil
				default:
					t.Fatalf("unexpected host: %s", r.URL.Host)
				}
				return nil, nil
			}),
		},
	})

	generated, err := generator.GenerateURL(context.Background(), "https://vxtwitter.com/example/status/1234567890?s=20")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.URL != "https://vxtwitter.com/example/status/1234567890?s=20" {
		t.Fatalf("unexpected normalized url: %q", generated.URL)
	}
	if generated.Type != "rich" {
		t.Fatalf("unexpected embed type: %q", generated.Type)
	}
	if generated.Author == nil || generated.Author.Name != "Example Display (@example)" || generated.Author.URL != "https://x.com/example/status/1234567890" {
		t.Fatalf("unexpected author: %#v", generated.Author)
	}
	if generated.Author.IconURL != "https://pbs.twimg.com/profile_images/example_normal.jpg" {
		t.Fatalf("unexpected author icon: %#v", generated.Author)
	}
	if generated.Description != "Proxy tweet text" {
		t.Fatalf("unexpected description: %q", generated.Description)
	}
	if generated.Timestamp == nil || generated.Timestamp.Unix() != 1700000000 {
		t.Fatalf("unexpected timestamp: %#v", generated.Timestamp)
	}
	if generated.Footer == nil || generated.Footer.Text != "vxTwitter / fixvx" {
		t.Fatalf("unexpected footer: %#v", generated.Footer)
	}
	if generated.Footer.IconURL != "https://vxtwitter.com/video.png" {
		t.Fatalf("unexpected footer icon: %#v", generated.Footer)
	}
	if generated.Thumbnail == nil || generated.Thumbnail.URL != "https://pbs.twimg.com/ext_tw_video_thumb/1234567890/pu/img/thumb.jpg" {
		t.Fatalf("unexpected thumbnail: %#v", generated.Thumbnail)
	}
	if generated.Thumbnail.Width == nil || *generated.Thumbnail.Width != 720 || generated.Thumbnail.Height == nil || *generated.Thumbnail.Height != 1280 {
		t.Fatalf("unexpected thumbnail dimensions: %#v", generated.Thumbnail)
	}
	if generated.Video == nil || generated.Video.URL != "https://vxtwitter.com/tvid/ext_tw_video/1234567890/pu/vid/avc1/720x1280/video" {
		t.Fatalf("unexpected video: %#v", generated.Video)
	}
	if generated.Video.ContentType != "video/mp4" {
		t.Fatalf("unexpected video content type: %#v", generated.Video)
	}
	if generated.Color == nil || *generated.Color != 0x1DA1F2 {
		t.Fatalf("unexpected twitter color: %#v", generated.Color)
	}
}

func TestGenerateFXTwitterEmbedUsesServiceAPI(t *testing.T) {
	generator := mustNewGenerator(t, Config{
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "api.fxtwitter.com" {
					t.Fatalf("unexpected host: %s", r.URL.Host)
				}
				if r.URL.Path != "/example/status/9988776655" {
					t.Fatalf("unexpected api path: %s", r.URL.Path)
				}

				body := `{
					"code": 200,
					"tweet": {
						"url": "https://twitter.com/example/status/9988776655",
						"id": "9988776655",
						"text": "Second proxy tweet",
						"author": {
							"name": "Example Two",
							"screen_name": "example",
							"avatar_url": "https://pbs.twimg.com/profile_images/example_two_normal.jpg"
						},
						"media": {
							"photos": [{
								"type": "photo",
								"url": "https://pbs.twimg.com/media/example.jpg",
								"width": 1200,
								"height": 675
							}]
						}
					}
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request:    r,
				}, nil
			}),
		},
	})

	generated, err := generator.GenerateURL(context.Background(), "https://fxtwitter.com/example/status/9988776655")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.URL != "https://fxtwitter.com/example/status/9988776655" {
		t.Fatalf("unexpected canonical url: %q", generated.URL)
	}
	if generated.Author == nil || generated.Author.Name != "Example Two (@example)" || generated.Author.URL != "https://twitter.com/example/status/9988776655" {
		t.Fatalf("unexpected author: %#v", generated.Author)
	}
	if generated.Description != "Second proxy tweet" {
		t.Fatalf("unexpected description: %q", generated.Description)
	}
	if generated.Footer == nil || generated.Footer.Text != "FixTweet / fxtwitter" {
		t.Fatalf("unexpected footer: %#v", generated.Footer)
	}
	if generated.Image == nil || generated.Image.URL != "https://pbs.twimg.com/media/example.jpg" {
		t.Fatalf("unexpected image: %#v", generated.Image)
	}
}

func TestGenerateXStatusURLUsesTwitterStatusAPI(t *testing.T) {
	generator := mustNewGenerator(t, Config{
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "api.fxtwitter.com" {
					t.Fatalf("unexpected host: %s", r.URL.Host)
				}
				if r.URL.Path != "/example/status/1122334455" {
					t.Fatalf("unexpected api path: %s", r.URL.Path)
				}
				body := `{
					"code": 200,
					"tweet": {
						"url": "https://twitter.com/example/status/1122334455",
						"id": "1122334455",
						"text": "Tweet from x.com",
						"author": {
							"name": "Example Three",
							"screen_name": "example"
						}
					}
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request:    r,
				}, nil
			}),
		},
	})

	generated, err := generator.GenerateURL(context.Background(), "https://x.com/example/status/1122334455")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.URL != "https://x.com/example/status/1122334455" {
		t.Fatalf("unexpected canonical url: %q", generated.URL)
	}
	if generated.Author == nil || generated.Author.Name != "Example Three (@example)" || generated.Author.URL != "https://twitter.com/example/status/1122334455" {
		t.Fatalf("unexpected author: %#v", generated.Author)
	}
	if generated.Description != "Tweet from x.com" {
		t.Fatalf("unexpected description: %q", generated.Description)
	}
	if generated.Footer != nil {
		t.Fatalf("expected no service footer for x.com, got %#v", generated.Footer)
	}
}

func TestGenerateTwitterStatusEmbedFallsBackToOEmbedDescription(t *testing.T) {
	generator := mustNewGenerator(t, Config{
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Host {
				case "api.fxtwitter.com":
					body := `{
						"code": 200,
						"tweet": {
							"url": "https://twitter.com/example/status/5566778899",
							"id": "5566778899",
							"text": "",
							"author": {
								"name": "Example Four",
								"screen_name": "example"
							}
						}
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
						Body:       io.NopCloser(bytes.NewBufferString(body)),
						Request:    r,
					}, nil
				case "publish.twitter.com":
					body := `{
						"type": "rich",
						"url": "https://twitter.com/example/status/5566778899",
						"html": "<blockquote class=\"twitter-tweet\"><p>Fallback tweet text</p></blockquote>"
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
						Body:       io.NopCloser(bytes.NewBufferString(body)),
						Request:    r,
					}, nil
				default:
					t.Fatalf("unexpected host: %s", r.URL.Host)
				}
				return nil, nil
			}),
		},
	})

	generated, err := generator.GenerateURL(context.Background(), "https://x.com/example/status/5566778899")
	if err != nil {
		t.Fatalf("GenerateURL returned error: %v", err)
	}
	if generated == nil {
		t.Fatal("expected embed")
	}
	if generated.Description != "Fallback tweet text" {
		t.Fatalf("unexpected description: %q", generated.Description)
	}
}

func TestGenerateSkipsTitleOnlyPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<!doctype html><html><head>
			<title>Only title</title>
			<meta property="og:site_name" content="Example Site">
		</head></html>`)
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})

	generated, err := generator.GenerateURL(context.Background(), server.URL)
	if !errors.Is(err, errSkipEmbed) {
		t.Fatalf("expected skip error, got %v", err)
	}
	if generated != nil {
		t.Fatalf("expected no embed, got %#v", generated)
	}

	embeds, err := generator.Generate(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 0 {
		t.Fatalf("expected no embeds, got %#v", embeds)
	}
}

func TestGenerateSkipsURLWhenPageCannotBeOpened(t *testing.T) {
	generator := mustNewGenerator(t, Config{
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			}),
		},
	})

	generated, err := generator.GenerateURL(context.Background(), "https://example.com/article")
	if !errors.Is(err, errSkipEmbed) {
		t.Fatalf("expected skip error, got %v", err)
	}
	if generated != nil {
		t.Fatalf("expected no embed, got %#v", generated)
	}

	embeds, err := generator.Generate(context.Background(), "https://example.com/article", nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 0 {
		t.Fatalf("expected no embeds, got %#v", embeds)
	}
}

func TestGenerateURLUsesCachedEmbed(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprint(w, `<!doctype html><html><head>
			<meta property="og:title" content="Example title">
			<meta property="og:description" content="Example description">
		</head></html>`)
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{
		AllowPrivateHosts: true,
		Cache:             newMemoryCache(),
	})

	first, err := generator.GenerateURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("first GenerateURL returned error: %v", err)
	}
	second, err := generator.GenerateURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("second GenerateURL returned error: %v", err)
	}
	if first == nil || second == nil {
		t.Fatalf("expected embeds, got %#v and %#v", first, second)
	}
	if hits != 1 {
		t.Fatalf("expected single origin fetch, got %d", hits)
	}
}

func TestGenerateURLUsesNegativeCacheForSkippedResult(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprint(w, `<!doctype html><html><head><title>Only title</title></head></html>`)
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{
		AllowPrivateHosts: true,
		Cache:             newMemoryCache(),
	})

	for i := 0; i < 2; i++ {
		generated, err := generator.GenerateURL(context.Background(), server.URL)
		if !errors.Is(err, errSkipEmbed) {
			t.Fatalf("expected skip error, got %v", err)
		}
		if generated != nil {
			t.Fatalf("expected no embed, got %#v", generated)
		}
	}
	if hits != 1 {
		t.Fatalf("expected single origin fetch for skipped URL, got %d", hits)
	}
}

func TestGenerateProbesOpenGraphImageDimensionsWhenMissing(t *testing.T) {
	imageHits := 0
	rangeHeader := ""
	preview := pngPayload(120, 63)

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/preview.png":
			imageHits++
			rangeHeader = r.Header.Get("Range")
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(preview)
		default:
			fmt.Fprintf(w, `<!doctype html><html><head>
				<meta property="og:title" content="Example title">
				<meta property="og:description" content="Example description">
				<meta property="og:image" content="%s/preview.png">
			</head></html>`, server.URL)
		}
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})
	embeds, err := generator.Generate(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embeds))
	}
	if embeds[0].Thumbnail == nil || embeds[0].Thumbnail.Width == nil || embeds[0].Thumbnail.Height == nil {
		t.Fatalf("expected probed thumbnail dimensions, got %#v", embeds[0].Thumbnail)
	}
	if *embeds[0].Thumbnail.Width != 120 || *embeds[0].Thumbnail.Height != 63 {
		t.Fatalf("unexpected probed thumbnail dimensions: %#v", embeds[0].Thumbnail)
	}
	if imageHits != 1 {
		t.Fatalf("expected single image probe, got %d", imageHits)
	}
	if rangeHeader != imageProbeRangeHeaderValue {
		t.Fatalf("unexpected range header: %q", rangeHeader)
	}
}

func TestGenerateSkipsOpenGraphImageProbeWhenDimensionsProvided(t *testing.T) {
	imageHits := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/preview.png":
			imageHits++
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngPayload(120, 63))
		default:
			fmt.Fprintf(w, `<!doctype html><html><head>
				<meta property="og:title" content="Example title">
				<meta property="og:description" content="Example description">
				<meta property="og:image" content="%s/preview.png">
				<meta property="og:image:width" content="120">
				<meta property="og:image:height" content="63">
			</head></html>`, server.URL)
		}
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{AllowPrivateHosts: true})
	embeds, err := generator.Generate(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embeds))
	}
	if embeds[0].Thumbnail == nil || embeds[0].Thumbnail.Width == nil || embeds[0].Thumbnail.Height == nil {
		t.Fatalf("expected metadata thumbnail dimensions, got %#v", embeds[0].Thumbnail)
	}
	if *embeds[0].Thumbnail.Width != 120 || *embeds[0].Thumbnail.Height != 63 {
		t.Fatalf("unexpected thumbnail dimensions: %#v", embeds[0].Thumbnail)
	}
	if imageHits != 0 {
		t.Fatalf("expected no image probe, got %d", imageHits)
	}
}

func TestGenerateURLSkipsExcludedPattern(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprint(w, `<!doctype html><html><head>
			<meta property="og:title" content="Example title">
			<meta property="og:description" content="Example description">
		</head></html>`)
	}))
	defer server.Close()

	generator := mustNewGenerator(t, Config{
		AllowPrivateHosts:   true,
		ExcludedURLPatterns: []string{`example title`, `^` + regexp.QuoteMeta(server.URL)},
	})

	generated, err := generator.GenerateURL(context.Background(), server.URL)
	if !errors.Is(err, errSkipEmbed) {
		t.Fatalf("expected skip error, got %v", err)
	}
	if generated != nil {
		t.Fatalf("expected no embed, got %#v", generated)
	}
	if hits != 0 {
		t.Fatalf("expected excluded URL to skip fetch, got %d hits", hits)
	}
}

func TestNewRejectsInvalidExcludedPattern(t *testing.T) {
	_, err := New(Config{
		ExcludedURLPatterns: []string{"["},
	})
	if err == nil {
		t.Fatal("expected invalid pattern error")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

type memoryCache struct {
	values map[string][]byte
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string][]byte)}
}

func (m *memoryCache) Set(ctx context.Context, key, val string) error {
	m.values[key] = []byte(val)
	return nil
}

func (m *memoryCache) Get(ctx context.Context, key string) (string, error) {
	value, ok := m.values[key]
	if !ok {
		return "", errors.New("missing key")
	}
	return string(value), nil
}

func (m *memoryCache) Delete(ctx context.Context, key string) error {
	delete(m.values, key)
	return nil
}

func (m *memoryCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	value, ok := m.values[key]
	if !ok {
		return nil, errors.New("missing key")
	}
	return append([]byte(nil), value...), nil
}

func (m *memoryCache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	return m.Set(ctx, key, val)
}

func (m *memoryCache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	return m.Set(ctx, key, fmt.Sprintf("%d", val))
}

func (m *memoryCache) SetInt64(ctx context.Context, key string, val int64) error {
	return m.Set(ctx, key, fmt.Sprintf("%d", val))
}

func (m *memoryCache) SetTTL(ctx context.Context, key string, ttl int64) error {
	return nil
}

func (m *memoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *memoryCache) GetInt64(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *memoryCache) SetJSON(ctx context.Context, key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	m.values[key] = data
	return nil
}

func (m *memoryCache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	return m.SetJSON(ctx, key, val)
}

func (m *memoryCache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	if _, ok := m.values[key]; ok {
		return false, nil
	}
	return true, m.SetJSON(ctx, key, val)
}

func (m *memoryCache) GetJSON(ctx context.Context, key string, v interface{}) error {
	value, ok := m.values[key]
	if !ok {
		return errors.New("missing key")
	}
	return json.Unmarshal(value, v)
}

func (m *memoryCache) HGet(ctx context.Context, key, field string) (string, error) {
	return "", nil
}

func (m *memoryCache) HSet(ctx context.Context, key, field, value string) error {
	return nil
}

func (m *memoryCache) HDel(ctx context.Context, key, field string) error {
	return nil
}

func (m *memoryCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (m *memoryCache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	return nil
}

func pngPayload(width, height int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	var encoded bytes.Buffer
	_ = png.Encode(&encoded, img)
	return encoded.Bytes()
}
