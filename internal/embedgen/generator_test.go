package embedgen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
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
