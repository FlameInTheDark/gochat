package embedgen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/embed"
)

const (
	defaultFetchTimeout              = 10 * time.Second
	defaultMaxBodyBytes        int64 = 2 << 20
	defaultUserAgent                 = "GoChat-Embedder/1.0"
	defaultYouTubeOEmbedURL          = "https://www.youtube.com/oembed"
	defaultYouTubeEmbedBaseURL       = "https://www.youtube.com/embed"
	defaultFixTweetAPIBaseURL        = "https://api.fxtwitter.com"
	defaultTwitterOEmbedURL          = "https://publish.twitter.com/oembed"
	defaultTwitterBaseURL            = "https://twitter.com"
	defaultCacheTTL                  = 6 * time.Hour
	defaultNegativeCacheTTL          = 30 * time.Minute
	youTubeBrandColor                = 0xFF0000
	twitterBrandColor                = 0x1DA1F2
)

var (
	embedURLRegex     = regexp.MustCompile(`(?i)\bhttps?://[^\s]+`)
	twitterHosts      = []string{"twitter.com", "x.com", "vxtwitter.com", "fxtwitter.com", "fixvx.com", "fixupx.com"}
	twitterProxyHosts = []string{"vxtwitter.com", "fxtwitter.com", "fixvx.com", "fixupx.com"}
	blockedPrefixes   = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("169.254.0.0/16"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("224.0.0.0/4"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("::/128"),
		netip.MustParsePrefix("::1/128"),
		netip.MustParsePrefix("fc00::/7"),
		netip.MustParsePrefix("fe80::/10"),
		netip.MustParsePrefix("ff00::/8"),
		netip.MustParsePrefix("2001:db8::/32"),
	}

	errSkipEmbed = errors.New("skip embed generation")
)

type Config struct {
	HTTPClient            *http.Client
	Cache                 cache.Cache
	CacheTTL              time.Duration
	NegativeCacheTTL      time.Duration
	ExcludedURLPatterns   []string
	AllowPrivateHosts     bool
	FetchTimeout          time.Duration
	MaxBodyBytes          int64
	UserAgent             string
	YouTubeOEmbedEndpoint string
	YouTubeEmbedBaseURL   string
}

type Generator struct {
	client                *http.Client
	cache                 cache.Cache
	cacheTTL              time.Duration
	negativeCacheTTL      time.Duration
	excludedURLPatterns   []*regexp.Regexp
	allowPrivateHosts     bool
	maxBodyBytes          int64
	userAgent             string
	youtubeOEmbedEndpoint string
	youtubeEmbedBaseURL   string
}

type fetchedPage struct {
	FinalURL    *url.URL
	ContentType string
	Body        []byte
}

type pageMetadata struct {
	HTMLTitle          string
	Description        string
	OGTitle            string
	OGDescription      string
	TwitterTitle       string
	TwitterDescription string
	SiteName           string
	CanonicalURL       string
	OEmbedURL          string
	OGType             string
	TwitterCard        string
	ImageURL           string
	ImageWidth         *int64
	ImageHeight        *int64
	VideoURL           string
	VideoWidth         *int64
	VideoHeight        *int64
	AuthorName         string
}

type oEmbedResponse struct {
	Type            string
	Title           string
	Description     string
	AuthorName      string
	AuthorURL       string
	ProviderName    string
	ProviderURL     string
	ThumbnailURL    string
	ThumbnailWidth  *int64
	ThumbnailHeight *int64
	Width           *int64
	Height          *int64
	URL             string
	HTML            string
}

type twitterStatusReference struct {
	Username         string
	StatusID         string
	CanonicalHost    string
	CanonicalURL     string
	AuthorURL        string
	AlternateService bool
}

type twitterStatusService struct {
	APIBaseURL        string
	FooterText        string
	FooterIconURL     string
	VideoFooterIcon   string
	VideoProxyBaseURL string
}

type twitterAPIStatusResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Tweet   *twitterAPIStatus `json:"tweet"`
}

type twitterAPIStatus struct {
	URL              string              `json:"url"`
	ID               string              `json:"id"`
	Text             string              `json:"text"`
	Color            string              `json:"color"`
	TwitterCard      string              `json:"twitter_card"`
	CreatedAt        string              `json:"created_at"`
	CreatedTimestamp *int64              `json:"created_timestamp"`
	Source           string              `json:"source"`
	Author           *twitterAPIAuthor   `json:"author"`
	Media            *twitterAPIMediaSet `json:"media"`
}

type twitterAPIAuthor struct {
	Name        string `json:"name"`
	ScreenName  string `json:"screen_name"`
	AvatarURL   string `json:"avatar_url"`
	AvatarColor string `json:"avatar_color"`
}

type twitterAPIMediaSet struct {
	External *twitterAPIExternalMedia `json:"external"`
	Photos   []twitterAPIPhoto        `json:"photos"`
	Videos   []twitterAPIVideo        `json:"videos"`
}

type twitterAPIExternalMedia struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Height   *int64 `json:"height"`
	Width    *int64 `json:"width"`
	Duration *int64 `json:"duration"`
}

type twitterAPIPhoto struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Height *int64 `json:"height"`
	Width  *int64 `json:"width"`
}

type twitterAPIVideo struct {
	Type         string `json:"type"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Height       *int64 `json:"height"`
	Width        *int64 `json:"width"`
	Format       string `json:"format"`
}

type cachedEmbedResult struct {
	Skip  bool         `json:"skip,omitempty"`
	Embed *embed.Embed `json:"embed,omitempty"`
}

func New(cfg Config) (*Generator, error) {
	fetchTimeout := cfg.FetchTimeout
	if fetchTimeout <= 0 {
		fetchTimeout = defaultFetchTimeout
	}

	maxBodyBytes := cfg.MaxBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxBodyBytes
	}

	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = defaultCacheTTL
	}

	negativeCacheTTL := cfg.NegativeCacheTTL
	if negativeCacheTTL <= 0 {
		negativeCacheTTL = defaultNegativeCacheTTL
	}

	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	youtubeOEmbedEndpoint := strings.TrimSpace(cfg.YouTubeOEmbedEndpoint)
	if youtubeOEmbedEndpoint == "" {
		youtubeOEmbedEndpoint = defaultYouTubeOEmbedURL
	}
	youtubeEmbedBaseURL := strings.TrimRight(strings.TrimSpace(cfg.YouTubeEmbedBaseURL), "/")
	if youtubeEmbedBaseURL == "" {
		youtubeEmbedBaseURL = defaultYouTubeEmbedBaseURL
	}

	excludedURLPatterns := make([]*regexp.Regexp, 0, len(cfg.ExcludedURLPatterns))
	for _, pattern := range cfg.ExcludedURLPatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid excluded url pattern %q: %w", pattern, err)
		}
		excludedURLPatterns = append(excludedURLPatterns, compiled)
	}

	g := &Generator{
		cache:                 cfg.Cache,
		cacheTTL:              cacheTTL,
		negativeCacheTTL:      negativeCacheTTL,
		excludedURLPatterns:   excludedURLPatterns,
		allowPrivateHosts:     cfg.AllowPrivateHosts,
		maxBodyBytes:          maxBodyBytes,
		userAgent:             userAgent,
		youtubeOEmbedEndpoint: youtubeOEmbedEndpoint,
		youtubeEmbedBaseURL:   youtubeEmbedBaseURL,
	}
	if cfg.HTTPClient != nil {
		g.client = cfg.HTTPClient
		return g, nil
	}

	transport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DialContext:       g.dialContext,
		ForceAttemptHTTP2: true,
	}
	g.client = &http.Client{
		Timeout:   fetchTimeout,
		Transport: transport,
	}
	return g, nil
}

func ExtractURLs(text string) []string {
	matches := embedURLRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	urls := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		match = strings.TrimRight(match, ".,!?;:)]}>\"'")
		if match == "" {
			continue
		}
		if _, ok := seen[match]; ok {
			continue
		}
		seen[match] = struct{}{}
		urls = append(urls, match)
	}
	if len(urls) == 0 {
		return nil
	}
	return urls
}

func (g *Generator) Generate(ctx context.Context, text string, manualEmbeds []embed.Embed) ([]embed.Embed, error) {
	remaining := embed.MaxEmbedsPerMessage - len(manualEmbeds)
	if remaining <= 0 {
		return nil, nil
	}

	urls := ExtractURLs(text)
	if len(urls) == 0 {
		return nil, nil
	}

	generated := make([]embed.Embed, 0, remaining)
	var firstErr error
	for _, rawURL := range urls {
		if len(generated) >= remaining {
			break
		}

		candidate, err := g.GenerateURL(ctx, rawURL)
		if err != nil {
			if errors.Is(err, errSkipEmbed) {
				continue
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if candidate == nil {
			continue
		}

		next := append(append([]embed.Embed(nil), generated...), *candidate)
		combined := embed.MergeEmbeds(manualEmbeds, next)
		if err := embed.ValidateEmbeds(combined); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		generated = next
	}

	if len(generated) == 0 {
		return nil, firstErr
	}
	return generated, nil
}

func (g *Generator) GenerateURL(ctx context.Context, rawURL string) (*embed.Embed, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !isHTTPURL(parsedURL) {
		return nil, skipEmbedError("invalid embed URL %q", rawURL)
	}
	twitterRef, hasTwitterRef := extractTwitterStatusReference(parsedURL)
	normalizedURL := parsedURL.String()
	if g.shouldExcludeURL(rawURL, normalizedURL) {
		return nil, skipEmbedError("excluded embed URL %s", normalizedURL)
	}

	if cached, cachedErr, ok := g.loadCachedResult(ctx, normalizedURL); ok {
		return cached, cachedErr
	}
	if hasTwitterRef {
		if result, err := g.generateTwitterStatusEmbed(ctx, parsedURL.String(), twitterRef); err == nil && result != nil {
			g.populateMissingEmbedMediaDimensions(ctx, result)
			if err := embed.ValidateEmbeds([]embed.Embed{*result}); err != nil {
				g.storeSkippedResult(ctx, normalizedURL)
				return nil, skipEmbedError("invalid embed metadata for %s", parsedURL.String())
			}
			g.storeEmbedResult(ctx, normalizedURL, result)
			return result, nil
		}
	}

	pageURL := parsedURL
	youtubeID, hasYouTubeID := extractYouTubeID(parsedURL)
	oembedURL := ""
	if hasYouTubeID {
		oembedURL = g.buildYouTubeOEmbedURL(parsedURL.String())
	} else if hasTwitterRef {
		oembedURL = buildTwitterOEmbedURL(firstNonEmpty(twitterRef.CanonicalURL, normalizedURL))
		if canonicalParsed, parseErr := url.Parse(firstNonEmpty(twitterRef.CanonicalURL, normalizedURL)); parseErr == nil && canonicalParsed != nil {
			pageURL = canonicalParsed
		}
	}

	var page fetchedPage
	var pageErr error
	if !hasTwitterRef {
		page, pageErr = g.fetch(ctx, parsedURL.String())
		if pageErr == nil {
			pageURL = page.FinalURL
			if page.FinalURL != nil {
				if id, ok := extractYouTubeID(page.FinalURL); ok {
					youtubeID = id
					hasYouTubeID = true
					if oembedURL == "" {
						oembedURL = g.buildYouTubeOEmbedURL(page.FinalURL.String())
					}
				}
			}

			if isImageContentType(page.ContentType) {
				embedURL := pageURL.String()
				var width *int64
				var height *int64
				if parsedWidth, parsedHeight, decodeErr := decodeImageDimensions(page.Body); decodeErr == nil {
					width = &parsedWidth
					height = &parsedHeight
				}
				result := &embed.Embed{
					Type: "image",
					URL:  embedURL,
					Image: &embed.EmbedMedia{
						URL:         embedURL,
						Width:       width,
						Height:      height,
						ContentType: page.ContentType,
					},
				}
				g.storeEmbedResult(ctx, normalizedURL, result)
				return result, nil
			}
		}
	}

	var metadata pageMetadata
	if !hasTwitterRef && pageErr == nil && (page.ContentType == "" || isHTMLContentType(page.ContentType)) {
		metadata, err = parseHTMLMetadata(pageURL, page.Body)
		if err != nil && oembedURL == "" {
			err = skipEmbedError("unable to parse metadata for %s", parsedURL.String())
			g.storeSkippedResult(ctx, normalizedURL)
			return nil, err
		}
		if oembedURL == "" {
			oembedURL = metadata.OEmbedURL
		}
	}

	var oembedData *oEmbedResponse
	if oembedURL != "" {
		oembedData, err = g.fetchOEmbed(ctx, oembedURL)
		if err != nil && pageErr != nil {
			err = skipEmbedError("unable to fetch page or oEmbed data for %s", parsedURL.String())
			g.storeSkippedResult(ctx, normalizedURL)
			return nil, err
		}
	}
	if pageErr != nil && oembedData == nil {
		err = skipEmbedError("unable to open %s", parsedURL.String())
		g.storeSkippedResult(ctx, normalizedURL)
		return nil, err
	}

	youtubeVideoURL := ""
	if hasYouTubeID {
		youtubeVideoURL = g.buildYouTubeVideoURL(youtubeID)
	}

	result := buildEmbed(parsedURL.String(), pageURL, metadata, oembedData, youtubeVideoURL)
	if result == nil {
		err = skipEmbedError("insufficient embed metadata for %s", parsedURL.String())
		g.storeSkippedResult(ctx, normalizedURL)
		return nil, err
	}
	g.populateMissingEmbedMediaDimensions(ctx, result)
	if err := embed.ValidateEmbeds([]embed.Embed{*result}); err != nil {
		g.storeSkippedResult(ctx, normalizedURL)
		return nil, skipEmbedError("invalid embed metadata for %s", parsedURL.String())
	}
	g.storeEmbedResult(ctx, normalizedURL, result)
	return result, nil
}

func (g *Generator) generateTwitterStatusEmbed(ctx context.Context, originalURL string, reference twitterStatusReference) (*embed.Embed, error) {
	service := twitterStatusServiceForHost(hostFromRawURL(originalURL))
	if service.APIBaseURL == "" {
		return nil, nil
	}

	endpoint := buildTwitterStatusAPIURL(service.APIBaseURL, reference.Username, reference.StatusID)
	if endpoint == "" {
		return nil, nil
	}

	status, err := g.fetchTwitterStatus(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(status.Text) == "" {
		oembedURL := buildTwitterOEmbedURL(firstNonEmpty(status.URL, reference.CanonicalURL))
		if oembedURL != "" {
			if oembedData, oembedErr := g.fetchOEmbed(ctx, oembedURL); oembedErr == nil && oembedData != nil {
				status.Text = firstNonEmpty(status.Text, oembedData.Description, descriptionFromOEmbedHTML(oembedData.HTML))
			}
		}
	}
	return buildTwitterStatusEmbed(originalURL, reference, service, status), nil
}

func buildEmbed(originalURL string, pageURL *url.URL, metadata pageMetadata, oembedData *oEmbedResponse, youtubeVideoURL string) *embed.Embed {
	embedURL := originalURL
	if metadata.CanonicalURL != "" {
		embedURL = metadata.CanonicalURL
	} else if valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.URL }) != "" {
		embedURL = valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.URL })
	} else if pageURL != nil {
		embedURL = pageURL.String()
	}

	originalParsed, _ := url.Parse(strings.TrimSpace(originalURL))
	embedParsed, _ := url.Parse(strings.TrimSpace(embedURL))
	twitterRef, hasTwitterRef := firstTwitterStatusReference(embedParsed, pageURL, originalParsed)
	if hasTwitterRef && twitterRef.AlternateService && urlHasKnownHost(embedURL, twitterProxyHosts...) {
		embedURL = twitterRef.CanonicalURL
	}

	title := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.Title }),
		metadata.OGTitle,
		metadata.TwitterTitle,
		metadata.HTMLTitle,
	), embed.MaxTitleCharacters)
	description := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.Description }),
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return descriptionFromOEmbedHTML(v.HTML) }),
		metadata.OGDescription,
		metadata.TwitterDescription,
		metadata.Description,
	), embed.MaxDescriptionCharacters)
	providerName := firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.ProviderName }),
		metadata.SiteName,
	)
	providerURL := valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.ProviderURL })
	if providerURL == "" && providerName != "" {
		providerURL = siteRoot(pageURL)
	}
	authorName := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.AuthorName }),
		metadata.AuthorName,
	), embed.MaxAuthorNameCharacters)
	authorURL := valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.AuthorURL })
	if hasTwitterRef {
		twitterHost := twitterRef.CanonicalHost
		if host := hostFromRawURL(authorURL); host != "" {
			twitterHost = host
		} else if host := hostFromRawURL(embedURL); host != "" {
			twitterHost = host
		}
		providerName = "Twitter"
		if providerURL == "" || urlHasKnownHost(providerURL, twitterProxyHosts...) {
			providerURL = twitterBaseURL(twitterHost)
		}
		if authorName == "" && twitterRef.Username != "" {
			authorName = "@" + twitterRef.Username
		}
		if authorURL == "" {
			authorURL = buildTwitterAuthorURL(twitterHost, twitterRef.Username)
		}
	}
	imageURL := firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.ThumbnailURL }),
		metadata.ImageURL,
	)
	imageWidth := firstNonNil(
		valueOrNil(oembedData, func(v *oEmbedResponse) *int64 { return v.ThumbnailWidth }),
		metadata.ImageWidth,
	)
	imageHeight := firstNonNil(
		valueOrNil(oembedData, func(v *oEmbedResponse) *int64 { return v.ThumbnailHeight }),
		metadata.ImageHeight,
	)
	videoURL := metadata.VideoURL
	if youtubeVideoURL != "" {
		videoURL = youtubeVideoURL
	}
	videoWidth := firstNonNil(
		valueOrNil(oembedData, func(v *oEmbedResponse) *int64 { return v.Width }),
		metadata.VideoWidth,
	)
	videoHeight := firstNonNil(
		valueOrNil(oembedData, func(v *oEmbedResponse) *int64 { return v.Height }),
		metadata.VideoHeight,
	)

	kind := detectEmbedType(oembedData, metadata, imageURL, videoURL)
	if kind == "" {
		kind = "link"
	}

	result := &embed.Embed{
		Type:        kind,
		URL:         embedURL,
		Title:       title,
		Description: description,
		Color:       providerColor(providerName, providerURL, pageURL, embedURL),
	}
	if providerName != "" || providerURL != "" {
		result.Provider = &embed.EmbedProvider{
			Name: providerName,
			URL:  providerURL,
		}
	}
	if authorName != "" {
		result.Author = &embed.EmbedAuthor{
			Name: authorName,
			URL:  authorURL,
		}
	}
	if imageURL != "" {
		media := &embed.EmbedMedia{
			URL:    imageURL,
			Width:  imageWidth,
			Height: imageHeight,
		}
		if kind == "image" {
			result.Image = media
		} else {
			result.Thumbnail = media
		}
	}
	if videoURL != "" {
		result.Video = &embed.EmbedMedia{
			URL:    videoURL,
			Width:  videoWidth,
			Height: videoHeight,
		}
	}

	if !hasRenderableContent(result) {
		return nil
	}
	return result
}

func hasRenderableContent(result *embed.Embed) bool {
	if result == nil {
		return false
	}
	if result.Description != "" {
		return true
	}
	if result.Image != nil || result.Thumbnail != nil || result.Video != nil {
		return true
	}
	if len(result.Fields) > 0 {
		return true
	}
	return false
}

func skipEmbedError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errSkipEmbed, fmt.Sprintf(format, args...))
}

func (g *Generator) shouldExcludeURL(rawURL, normalizedURL string) bool {
	if len(g.excludedURLPatterns) == 0 {
		return false
	}
	for _, pattern := range g.excludedURLPatterns {
		if pattern == nil {
			continue
		}
		if pattern.MatchString(rawURL) || pattern.MatchString(normalizedURL) {
			return true
		}
	}
	return false
}

func (g *Generator) loadCachedResult(ctx context.Context, rawURL string) (*embed.Embed, error, bool) {
	if g.cache == nil || rawURL == "" {
		return nil, nil, false
	}

	var cached cachedEmbedResult
	if err := g.cache.GetJSON(ctx, embedCacheKey(rawURL), &cached); err != nil {
		return nil, nil, false
	}
	if cached.Skip {
		return nil, skipEmbedError("cached skip for %s", rawURL), true
	}
	if cached.Embed == nil {
		_ = g.cache.Delete(ctx, embedCacheKey(rawURL))
		return nil, nil, false
	}
	if err := embed.ValidateEmbeds([]embed.Embed{*cached.Embed}); err != nil {
		_ = g.cache.Delete(ctx, embedCacheKey(rawURL))
		return nil, nil, false
	}
	return cached.Embed, nil, true
}

func (g *Generator) storeEmbedResult(ctx context.Context, rawURL string, result *embed.Embed) {
	if g.cache == nil || rawURL == "" || result == nil || g.cacheTTL <= 0 {
		return
	}
	_ = g.cache.SetTimedJSON(ctx, embedCacheKey(rawURL), cachedEmbedResult{Embed: result}, int64(g.cacheTTL/time.Second))
}

func (g *Generator) storeSkippedResult(ctx context.Context, rawURL string) {
	if g.cache == nil || rawURL == "" || g.negativeCacheTTL <= 0 {
		return
	}
	_ = g.cache.SetTimedJSON(ctx, embedCacheKey(rawURL), cachedEmbedResult{Skip: true}, int64(g.negativeCacheTTL/time.Second))
}

func embedCacheKey(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return "embedder:url:v3:" + hex.EncodeToString(sum[:])
}

func providerColor(providerName, providerURL string, pageURL *url.URL, embedURL string) *int {
	if matchesYouTubeProvider(providerName, providerURL, pageURL, embedURL) {
		return intPtr(youTubeBrandColor)
	}
	if matchesTwitterProvider(providerName, providerURL, pageURL, embedURL) {
		return intPtr(twitterBrandColor)
	}
	return nil
}

func matchesYouTubeProvider(providerName, providerURL string, pageURL *url.URL, embedURL string) bool {
	if strings.Contains(strings.ToLower(strings.TrimSpace(providerName)), "youtube") {
		return true
	}
	return hasKnownHost(providerURL, pageURL, embedURL, "youtube.com", "youtu.be", "youtube-nocookie.com")
}

func matchesTwitterProvider(providerName, providerURL string, pageURL *url.URL, embedURL string) bool {
	if strings.Contains(strings.ToLower(strings.TrimSpace(providerName)), "twitter") {
		return true
	}
	return hasKnownHost(providerURL, pageURL, embedURL, "twitter.com", "x.com")
}

func hasKnownHost(providerURL string, pageURL *url.URL, embedURL string, hosts ...string) bool {
	candidates := make([]string, 0, 2)
	if providerURL != "" {
		candidates = append(candidates, providerURL)
	}
	if embedURL != "" {
		candidates = append(candidates, embedURL)
	}
	for _, candidate := range candidates {
		if urlHasKnownHost(candidate, hosts...) {
			return true
		}
	}
	if pageURL != nil {
		if hostMatches(pageURL.Hostname(), hosts...) {
			return true
		}
	}
	return false
}

func urlHasKnownHost(rawURL string, hosts ...string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return hostMatches(parsed.Hostname(), hosts...)
}

func hostMatches(rawHost string, hosts ...string) bool {
	host := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(rawHost)), "www.")
	if host == "" {
		return false
	}
	for _, candidate := range hosts {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		if host == candidate || strings.HasSuffix(host, "."+candidate) {
			return true
		}
	}
	return false
}

func intPtr(value int) *int {
	v := value
	return &v
}

func detectEmbedType(oembedData *oEmbedResponse, metadata pageMetadata, imageURL, videoURL string) string {
	oembedType := strings.ToLower(strings.TrimSpace(valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.Type })))
	switch oembedType {
	case "photo":
		return "image"
	case "video":
		return "video"
	case "link", "rich":
		return oembedType
	}

	ogType := strings.ToLower(strings.TrimSpace(metadata.OGType))
	twitterCard := strings.ToLower(strings.TrimSpace(metadata.TwitterCard))
	switch {
	case videoURL != "":
		return "video"
	case strings.HasPrefix(ogType, "video"):
		return "video"
	case twitterCard == "player":
		return "video"
	case ogType == "image":
		return "image"
	case imageURL != "" && ogType == "article":
		return "article"
	case ogType == "article":
		return "article"
	case imageURL != "":
		return "link"
	case metadata.OGTitle != "" || metadata.TwitterTitle != "" || metadata.Description != "" || metadata.OGDescription != "" || metadata.TwitterDescription != "":
		return "link"
	default:
		return ""
	}
}

func parseHTMLMetadata(baseURL *url.URL, body []byte) (pageMetadata, error) {
	var metadata pageMetadata
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return metadata, fmt.Errorf("unable to parse html metadata: %w", err)
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode {
			switch node.Data {
			case "title":
				if metadata.HTMLTitle == "" {
					metadata.HTMLTitle = strings.TrimSpace(textContent(node))
				}
			case "meta":
				attrs := attrMap(node)
				content := strings.TrimSpace(attrs["content"])
				if content != "" {
					applyMetaTag(&metadata, baseURL, attrs, content)
				}
			case "link":
				attrs := attrMap(node)
				applyLinkTag(&metadata, baseURL, attrs)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return metadata, nil
}

func applyMetaTag(metadata *pageMetadata, baseURL *url.URL, attrs map[string]string, content string) {
	key := strings.ToLower(strings.TrimSpace(attrs["property"]))
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(attrs["name"]))
	}
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(attrs["itemprop"]))
	}
	if key == "" {
		return
	}

	switch key {
	case "title":
		if metadata.HTMLTitle == "" {
			metadata.HTMLTitle = content
		}
	case "description":
		if metadata.Description == "" {
			metadata.Description = content
		}
	case "author", "article:author":
		if metadata.AuthorName == "" {
			metadata.AuthorName = content
		}
	case "og:title":
		if metadata.OGTitle == "" {
			metadata.OGTitle = content
		}
	case "og:description":
		if metadata.OGDescription == "" {
			metadata.OGDescription = content
		}
	case "og:site_name":
		if metadata.SiteName == "" {
			metadata.SiteName = content
		}
	case "og:type":
		if metadata.OGType == "" {
			metadata.OGType = content
		}
	case "og:url", "twitter:url":
		if metadata.CanonicalURL == "" {
			metadata.CanonicalURL = resolveURL(baseURL, content)
		}
	case "og:image", "og:image:url", "og:image:secure_url", "twitter:image", "twitter:image:src":
		if metadata.ImageURL == "" {
			metadata.ImageURL = resolveURL(baseURL, content)
		}
	case "og:image:width", "twitter:image:width":
		if metadata.ImageWidth == nil {
			metadata.ImageWidth = parseInt64String(content)
		}
	case "og:image:height", "twitter:image:height":
		if metadata.ImageHeight == nil {
			metadata.ImageHeight = parseInt64String(content)
		}
	case "og:video", "og:video:url", "og:video:secure_url", "twitter:player":
		if metadata.VideoURL == "" {
			metadata.VideoURL = resolveURL(baseURL, content)
		}
	case "og:video:width", "twitter:player:width":
		if metadata.VideoWidth == nil {
			metadata.VideoWidth = parseInt64String(content)
		}
	case "og:video:height", "twitter:player:height":
		if metadata.VideoHeight == nil {
			metadata.VideoHeight = parseInt64String(content)
		}
	case "twitter:title":
		if metadata.TwitterTitle == "" {
			metadata.TwitterTitle = content
		}
	case "twitter:description":
		if metadata.TwitterDescription == "" {
			metadata.TwitterDescription = content
		}
	case "twitter:creator":
		if metadata.AuthorName == "" {
			metadata.AuthorName = content
		}
	case "twitter:card":
		if metadata.TwitterCard == "" {
			metadata.TwitterCard = content
		}
	}
}

func applyLinkTag(metadata *pageMetadata, baseURL *url.URL, attrs map[string]string) {
	href := strings.TrimSpace(attrs["href"])
	if href == "" {
		return
	}
	rels := strings.Fields(strings.ToLower(attrs["rel"]))
	typ := strings.ToLower(strings.TrimSpace(attrs["type"]))
	for _, rel := range rels {
		switch rel {
		case "canonical":
			if metadata.CanonicalURL == "" {
				metadata.CanonicalURL = resolveURL(baseURL, href)
			}
		case "alternate":
			if metadata.OEmbedURL == "" && strings.Contains(typ, "oembed") {
				metadata.OEmbedURL = resolveURL(baseURL, href)
			}
		}
	}
}

func attrMap(node *html.Node) map[string]string {
	attrs := make(map[string]string, len(node.Attr))
	for _, attr := range node.Attr {
		attrs[strings.ToLower(attr.Key)] = attr.Val
	}
	return attrs
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return builder.String()
}

func (g *Generator) fetch(ctx context.Context, rawURL string) (fetchedPage, error) {
	var page fetchedPage
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return page, err
	}
	request.Header.Set("User-Agent", g.userAgent)

	response, err := g.client.Do(request)
	if err != nil {
		return page, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return page, fmt.Errorf("unexpected status code %d for %s", response.StatusCode, rawURL)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, g.maxBodyBytes+1))
	if err != nil {
		return page, err
	}
	if int64(len(body)) > g.maxBodyBytes {
		return page, fmt.Errorf("response body for %s exceeded %d bytes", rawURL, g.maxBodyBytes)
	}

	page.FinalURL = response.Request.URL
	page.ContentType = normalizeContentType(response.Header.Get("Content-Type"))
	page.Body = body
	return page, nil
}

func (g *Generator) fetchOEmbed(ctx context.Context, rawURL string) (*oEmbedResponse, error) {
	page, err := g.fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	data := make(map[string]json.RawMessage)
	if err := json.Unmarshal(page.Body, &data); err != nil {
		return nil, fmt.Errorf("unable to decode oEmbed payload: %w", err)
	}

	response := &oEmbedResponse{
		Type:            jsonString(data["type"]),
		Title:           jsonString(data["title"]),
		Description:     jsonString(data["description"]),
		AuthorName:      jsonString(data["author_name"]),
		AuthorURL:       jsonString(data["author_url"]),
		ProviderName:    jsonString(data["provider_name"]),
		ProviderURL:     jsonString(data["provider_url"]),
		ThumbnailURL:    jsonString(data["thumbnail_url"]),
		ThumbnailWidth:  jsonInt64(data["thumbnail_width"]),
		ThumbnailHeight: jsonInt64(data["thumbnail_height"]),
		Width:           jsonInt64(data["width"]),
		Height:          jsonInt64(data["height"]),
		URL:             jsonString(data["url"]),
		HTML:            jsonString(data["html"]),
	}
	return response, nil
}

func (g *Generator) fetchTwitterStatus(ctx context.Context, rawURL string) (*twitterAPIStatus, error) {
	page, err := g.fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	var response twitterAPIStatusResponse
	if err := json.Unmarshal(page.Body, &response); err != nil {
		return nil, fmt.Errorf("unable to decode twitter status payload: %w", err)
	}
	if response.Code != 0 && response.Code != http.StatusOK {
		return nil, fmt.Errorf("unexpected twitter status response code %d", response.Code)
	}
	if response.Tweet == nil {
		return nil, fmt.Errorf("twitter status payload did not include tweet data")
	}
	return response.Tweet, nil
}

func buildTwitterStatusEmbed(originalURL string, reference twitterStatusReference, service twitterStatusService, status *twitterAPIStatus) *embed.Embed {
	if status == nil {
		return nil
	}

	embedURL := firstNonEmpty(originalURL, status.URL, reference.CanonicalURL)
	description := truncateText(firstNonEmpty(status.Text), embed.MaxDescriptionCharacters)
	authorName := formatTwitterAuthorName(status.Author)
	authorURL := firstNonEmpty(status.URL, reference.CanonicalURL)
	authorIconURL := ""
	if status.Author != nil {
		authorIconURL = firstNonEmpty(status.Author.AvatarURL)
	}

	var timestamp *time.Time
	if status.CreatedTimestamp != nil {
		createdAt := time.Unix(*status.CreatedTimestamp, 0).UTC()
		timestamp = &createdAt
	} else if parsed := parseTwitterTimestamp(status.CreatedAt); parsed != nil {
		timestamp = parsed
	}

	var image *embed.EmbedMedia
	var thumbnail *embed.EmbedMedia
	var video *embed.EmbedMedia
	if status.Media != nil {
		if len(status.Media.Videos) > 0 {
			videoData := status.Media.Videos[0]
			videoURL := firstNonEmpty(videoData.URL)
			if proxiedURL := proxyTwitterVideoURL(service.VideoProxyBaseURL, videoURL); proxiedURL != "" {
				videoURL = proxiedURL
			}
			video = &embed.EmbedMedia{
				URL:         videoURL,
				Width:       firstNonNil(videoData.Width),
				Height:      firstNonNil(videoData.Height),
				ContentType: firstNonEmpty(videoData.Format),
			}
			if thumbURL := firstNonEmpty(videoData.ThumbnailURL); thumbURL != "" {
				thumbnail = &embed.EmbedMedia{
					URL:    thumbURL,
					Width:  firstNonNil(videoData.Width),
					Height: firstNonNil(videoData.Height),
				}
			}
		} else if len(status.Media.Photos) > 0 {
			photo := status.Media.Photos[0]
			image = &embed.EmbedMedia{
				URL:    firstNonEmpty(photo.URL),
				Width:  firstNonNil(photo.Width),
				Height: firstNonNil(photo.Height),
			}
		} else if status.Media.External != nil {
			external := status.Media.External
			video = &embed.EmbedMedia{
				URL:    firstNonEmpty(external.URL),
				Width:  firstNonNil(external.Width),
				Height: firstNonNil(external.Height),
			}
		}
	}

	result := &embed.Embed{
		Type:        "rich",
		URL:         embedURL,
		Description: description,
		Timestamp:   timestamp,
		Color:       intPtr(twitterBrandColor),
	}
	if authorName != "" || authorURL != "" || authorIconURL != "" {
		result.Author = &embed.EmbedAuthor{
			Name:    authorName,
			URL:     authorURL,
			IconURL: authorIconURL,
		}
	}
	if service.FooterText != "" || service.FooterIconURL != "" {
		footerIconURL := service.FooterIconURL
		if video != nil && video.URL != "" && service.VideoFooterIcon != "" {
			footerIconURL = service.VideoFooterIcon
		}
		result.Footer = &embed.EmbedFooter{
			Text:    service.FooterText,
			IconURL: footerIconURL,
		}
	}
	if image != nil && image.URL != "" {
		result.Image = image
	}
	if thumbnail != nil && thumbnail.URL != "" {
		result.Thumbnail = thumbnail
	}
	if video != nil && video.URL != "" {
		result.Video = video
	}
	if !hasRenderableContent(result) {
		return nil
	}
	return result
}

func buildTwitterStatusAPIURL(baseURL, username, statusID string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	statusID = strings.TrimSpace(statusID)
	if baseURL == "" || statusID == "" {
		return ""
	}
	pathParts := []string{"i", "status", statusID}
	if username != "" {
		pathParts = []string{username, "status", statusID}
	}
	return baseURL + "/" + strings.Join(pathParts, "/")
}

func twitterStatusServiceForHost(host string) twitterStatusService {
	switch {
	case hostMatches(host, "vxtwitter.com", "fixvx.com"):
		return twitterStatusService{
			APIBaseURL:        defaultFixTweetAPIBaseURL,
			FooterText:        "vxTwitter / fixvx",
			FooterIconURL:     "https://vxtwitter.com/favicon.ico",
			VideoFooterIcon:   "https://vxtwitter.com/video.png",
			VideoProxyBaseURL: "https://vxtwitter.com",
		}
	case hostMatches(host, "fxtwitter.com", "fixupx.com", "twittpr.com"):
		return twitterStatusService{
			APIBaseURL:    defaultFixTweetAPIBaseURL,
			FooterText:    "FixTweet / fxtwitter",
			FooterIconURL: "https://fxtwitter.com/favicon.ico",
		}
	default:
		return twitterStatusService{
			APIBaseURL: defaultFixTweetAPIBaseURL,
		}
	}
}

func formatTwitterAuthorName(author *twitterAPIAuthor) string {
	if author == nil {
		return ""
	}
	name := strings.TrimSpace(author.Name)
	screenName := strings.Trim(strings.TrimSpace(author.ScreenName), "@")
	switch {
	case name != "" && screenName != "":
		return truncateText(fmt.Sprintf("%s (@%s)", name, screenName), embed.MaxAuthorNameCharacters)
	case name != "":
		return truncateText(name, embed.MaxAuthorNameCharacters)
	case screenName != "":
		return truncateText("@"+screenName, embed.MaxAuthorNameCharacters)
	default:
		return ""
	}
}

func parseTwitterTimestamp(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"Mon Jan 02 15:04:05 -0700 2006",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			value := parsed.UTC()
			return &value
		}
	}
	return nil
}

func buildTwitterOEmbedURL(rawURL string) string {
	query := url.Values{}
	query.Set("url", rawURL)
	query.Set("omit_script", "true")
	return fmt.Sprintf("%s?%s", strings.TrimRight(defaultTwitterOEmbedURL, "?"), query.Encode())
}

func proxyTwitterVideoURL(baseURL, rawURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	rawURL = strings.TrimSpace(rawURL)
	if baseURL == "" || rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || !hostMatches(parsed.Hostname(), "video.twimg.com") {
		return ""
	}

	path := strings.Trim(parsed.EscapedPath(), "/")
	if path == "" {
		return ""
	}

	if lastSlash := strings.LastIndex(path, "/"); lastSlash >= 0 {
		filename := path[lastSlash+1:]
		if extension := strings.LastIndex(filename, "."); extension > 0 {
			path = path[:lastSlash+1] + filename[:extension]
		}
	} else if extension := strings.LastIndex(path, "."); extension > 0 {
		path = path[:extension]
	}

	return baseURL + "/tvid/" + path
}

func (g *Generator) buildYouTubeOEmbedURL(rawURL string) string {
	query := url.Values{}
	query.Set("url", rawURL)
	query.Set("format", "json")
	return fmt.Sprintf("%s?%s", strings.TrimRight(g.youtubeOEmbedEndpoint, "?"), query.Encode())
}

func (g *Generator) buildYouTubeVideoURL(videoID string) string {
	return fmt.Sprintf("%s/%s", g.youtubeEmbedBaseURL, videoID)
}

func descriptionFromOEmbedHTML(rawHTML string) string {
	rawHTML = strings.TrimSpace(rawHTML)
	if rawHTML == "" {
		return ""
	}

	root, err := html.Parse(strings.NewReader("<html><body>" + rawHTML + "</body></html>"))
	if err != nil {
		return ""
	}

	if paragraph := findNodeByTag(root, "p"); paragraph != nil {
		if text := normalizeInlineText(textContent(paragraph)); text != "" {
			return text
		}
	}

	return normalizeInlineText(textContent(root))
}

func findNodeByTag(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, tag) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findNodeByTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func normalizeInlineText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func (g *Generator) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if err := g.validateRemoteHost(ctx, host); err != nil {
		return nil, err
	}
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, net.JoinHostPort(host, port))
}

func (g *Generator) validateRemoteHost(ctx context.Context, host string) error {
	if g.allowPrivateHosts {
		return nil
	}

	normalizedHost := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if normalizedHost == "" {
		return fmt.Errorf("invalid empty host")
	}
	if normalizedHost == "localhost" || strings.HasSuffix(normalizedHost, ".local") {
		return fmt.Errorf("refusing to fetch private host %s", host)
	}
	if ip := net.ParseIP(normalizedHost); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("refusing to fetch private address %s", host)
		}
		return nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, normalizedHost)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return fmt.Errorf("host %s resolved to no addresses", host)
	}
	for _, addr := range addrs {
		if !isPublicIP(addr.IP) {
			return fmt.Errorf("refusing to fetch private address %s for host %s", addr.IP.String(), host)
		}
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return addr.IsValid() && addr.IsGlobalUnicast()
}

func extractYouTubeID(rawURL *url.URL) (string, bool) {
	if rawURL == nil {
		return "", false
	}
	host := strings.TrimPrefix(strings.ToLower(rawURL.Hostname()), "www.")
	switch host {
	case "youtube.com", "m.youtube.com":
		path := strings.Trim(rawURL.EscapedPath(), "/")
		switch {
		case path == "watch":
			id := strings.TrimSpace(rawURL.Query().Get("v"))
			return sanitizeYouTubeID(id)
		case strings.HasPrefix(path, "shorts/"):
			return sanitizeYouTubeID(strings.TrimPrefix(path, "shorts/"))
		case strings.HasPrefix(path, "embed/"):
			return sanitizeYouTubeID(strings.TrimPrefix(path, "embed/"))
		}
	case "youtu.be":
		return sanitizeYouTubeID(strings.Trim(strings.TrimSpace(rawURL.EscapedPath()), "/"))
	}
	return "", false
}

func firstTwitterStatusReference(urls ...*url.URL) (twitterStatusReference, bool) {
	for _, candidate := range urls {
		if reference, ok := extractTwitterStatusReference(candidate); ok {
			return reference, true
		}
	}
	return twitterStatusReference{}, false
}

func extractTwitterStatusReference(rawURL *url.URL) (twitterStatusReference, bool) {
	if rawURL == nil || !hostMatches(rawURL.Hostname(), twitterHosts...) {
		return twitterStatusReference{}, false
	}

	segments := splitPathSegments(rawURL.EscapedPath())
	if len(segments) < 2 {
		return twitterStatusReference{}, false
	}

	var username string
	var statusID string

	switch {
	case len(segments) >= 4 && strings.EqualFold(segments[0], "i") && strings.EqualFold(segments[1], "web") && strings.EqualFold(segments[2], "status"):
		statusID, _ = sanitizeTwitterStatusID(segments[3])
	case len(segments) >= 3 && strings.EqualFold(segments[0], "i") && strings.EqualFold(segments[1], "status"):
		statusID, _ = sanitizeTwitterStatusID(segments[2])
	case len(segments) >= 3 && strings.EqualFold(segments[1], "status"):
		username, _ = sanitizeTwitterScreenName(segments[0])
		statusID, _ = sanitizeTwitterStatusID(segments[2])
	default:
		return twitterStatusReference{}, false
	}

	if statusID == "" {
		return twitterStatusReference{}, false
	}

	canonicalHost := canonicalTwitterHost(rawURL.Hostname())
	return twitterStatusReference{
		Username:         username,
		StatusID:         statusID,
		CanonicalHost:    canonicalHost,
		CanonicalURL:     buildTwitterStatusURL(canonicalHost, username, statusID),
		AuthorURL:        buildTwitterAuthorURL(canonicalHost, username),
		AlternateService: hostMatches(rawURL.Hostname(), twitterProxyHosts...),
	}, true
}

func sanitizeTwitterScreenName(raw string) (string, bool) {
	raw = strings.Trim(strings.TrimSpace(raw), "@")
	if raw == "" {
		return "", false
	}
	for _, char := range raw {
		switch {
		case char == '_':
		case char >= '0' && char <= '9':
		case char >= 'A' && char <= 'Z':
		case char >= 'a' && char <= 'z':
		default:
			return "", false
		}
	}
	return raw, true
}

func sanitizeTwitterStatusID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if idx := strings.IndexAny(raw, "?/#&"); idx >= 0 {
		raw = raw[:idx]
	}
	for _, char := range raw {
		if char < '0' || char > '9' {
			return "", false
		}
	}
	return raw, true
}

func canonicalTwitterHost(rawHost string) string {
	switch {
	case hostMatches(rawHost, "x.com"):
		return "x.com"
	default:
		return "twitter.com"
	}
}

func buildTwitterStatusURL(host, username, statusID string) string {
	if host == "" || statusID == "" {
		return ""
	}
	pathParts := []string{"i", "status", statusID}
	if username != "" {
		pathParts = []string{username, "status", statusID}
	}
	return (&url.URL{Scheme: "https", Host: host, Path: "/" + strings.Join(pathParts, "/")}).String()
}

func buildTwitterAuthorURL(host, username string) string {
	if host == "" || username == "" {
		return ""
	}
	return (&url.URL{Scheme: "https", Host: host, Path: "/" + username}).String()
}

func twitterBaseURL(host string) string {
	if host == "" {
		return defaultTwitterBaseURL
	}
	return (&url.URL{Scheme: "https", Host: host}).String()
}

func hostFromRawURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func splitPathSegments(rawPath string) []string {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return nil
	}
	segments := strings.Split(strings.Trim(rawPath, "/"), "/")
	result := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		result = append(result, segment)
	}
	return result
}

func sanitizeYouTubeID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if idx := strings.IndexAny(raw, "?/&"); idx >= 0 {
		raw = raw[:idx]
	}
	return raw, raw != ""
}

func isHTTPURL(rawURL *url.URL) bool {
	if rawURL == nil {
		return false
	}
	scheme := strings.ToLower(strings.TrimSpace(rawURL.Scheme))
	return (scheme == "http" || scheme == "https") && rawURL.Host != ""
}

func resolveURL(baseURL *url.URL, rawValue string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawValue))
	if err != nil {
		return strings.TrimSpace(rawValue)
	}
	if baseURL == nil {
		return parsed.String()
	}
	return baseURL.ResolveReference(parsed).String()
}

func siteRoot(rawURL *url.URL) string {
	if rawURL == nil || rawURL.Scheme == "" || rawURL.Host == "" {
		return ""
	}
	return (&url.URL{Scheme: rawURL.Scheme, Host: rawURL.Host}).String()
}

func normalizeContentType(contentType string) string {
	if contentType == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(contentType))
	}
	return strings.ToLower(mediaType)
}

func isImageContentType(contentType string) bool {
	return strings.HasPrefix(normalizeContentType(contentType), "image/")
}

func isHTMLContentType(contentType string) bool {
	mediaType := normalizeContentType(contentType)
	return mediaType == "text/html" || mediaType == "application/xhtml+xml"
}

func parseInt64String(value string) *int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func jsonString(value json.RawMessage) string {
	if len(value) == 0 {
		return ""
	}
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err == nil {
		return strings.TrimSpace(stringValue)
	}
	return ""
}

func jsonInt64(value json.RawMessage) *int64 {
	if len(value) == 0 {
		return nil
	}
	var intValue int64
	if err := json.Unmarshal(value, &intValue); err == nil {
		return &intValue
	}
	var floatValue float64
	if err := json.Unmarshal(value, &floatValue); err == nil {
		intValue = int64(floatValue)
		return &intValue
	}
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err == nil {
		return parseInt64String(stringValue)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonNil(values ...*int64) *int64 {
	for _, value := range values {
		if value != nil {
			copyValue := *value
			return &copyValue
		}
	}
	return nil
}

func truncateText(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit]))
}

func valueOrEmpty[T any](value *T, fn func(*T) string) string {
	if value == nil {
		return ""
	}
	return fn(value)
}

func valueOrNil[T any](value *T, fn func(*T) *int64) *int64 {
	if value == nil {
		return nil
	}
	return fn(value)
}
