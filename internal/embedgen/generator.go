package embedgen

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/FlameInTheDark/gochat/internal/embed"
)

const (
	defaultFetchTimeout              = 10 * time.Second
	defaultMaxBodyBytes        int64 = 2 << 20
	defaultUserAgent                 = "GoChat-Embedder/1.0"
	defaultYouTubeOEmbedURL          = "https://www.youtube.com/oembed"
	defaultYouTubeEmbedBaseURL       = "https://www.youtube.com/embed"
	youTubeBrandColor                = 0xFF0000
	twitterBrandColor                = 0x1DA1F2
)

var (
	embedURLRegex   = regexp.MustCompile(`(?i)\bhttps?://[^\s]+`)
	blockedPrefixes = []netip.Prefix{
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
)

type Config struct {
	HTTPClient            *http.Client
	AllowPrivateHosts     bool
	FetchTimeout          time.Duration
	MaxBodyBytes          int64
	UserAgent             string
	YouTubeOEmbedEndpoint string
	YouTubeEmbedBaseURL   string
}

type Generator struct {
	client                *http.Client
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

func New(cfg Config) *Generator {
	fetchTimeout := cfg.FetchTimeout
	if fetchTimeout <= 0 {
		fetchTimeout = defaultFetchTimeout
	}

	maxBodyBytes := cfg.MaxBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxBodyBytes
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

	g := &Generator{
		allowPrivateHosts:     cfg.AllowPrivateHosts,
		maxBodyBytes:          maxBodyBytes,
		userAgent:             userAgent,
		youtubeOEmbedEndpoint: youtubeOEmbedEndpoint,
		youtubeEmbedBaseURL:   youtubeEmbedBaseURL,
	}
	if cfg.HTTPClient != nil {
		g.client = cfg.HTTPClient
		return g
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
	return g
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
		return nil, fmt.Errorf("invalid embed URL %q", rawURL)
	}

	pageURL := parsedURL
	youtubeID, hasYouTubeID := extractYouTubeID(parsedURL)
	oembedURL := ""
	if hasYouTubeID {
		oembedURL = g.buildYouTubeOEmbedURL(parsedURL.String())
	}

	page, pageErr := g.fetch(ctx, parsedURL.String())
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
			return &embed.Embed{
				Type: "image",
				URL:  embedURL,
				Image: &embed.EmbedMedia{
					URL:         embedURL,
					ContentType: page.ContentType,
				},
			}, nil
		}
	}

	var metadata pageMetadata
	if pageErr == nil && (page.ContentType == "" || isHTMLContentType(page.ContentType)) {
		metadata, err = parseHTMLMetadata(pageURL, page.Body)
		if err != nil && oembedURL == "" {
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
			return nil, err
		}
	}

	youtubeVideoURL := ""
	if hasYouTubeID {
		youtubeVideoURL = g.buildYouTubeVideoURL(youtubeID)
	}

	result := buildEmbed(parsedURL.String(), pageURL, metadata, oembedData, youtubeVideoURL)
	if result == nil {
		if pageErr != nil {
			return nil, pageErr
		}
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("no embed metadata found for %s", parsedURL.String())
	}
	if err := embed.ValidateEmbeds([]embed.Embed{*result}); err != nil {
		return nil, err
	}
	return result, nil
}

func buildEmbed(originalURL string, pageURL *url.URL, metadata pageMetadata, oembedData *oEmbedResponse, youtubeVideoURL string) *embed.Embed {
	embedURL := originalURL
	if metadata.CanonicalURL != "" {
		embedURL = metadata.CanonicalURL
	} else if pageURL != nil {
		embedURL = pageURL.String()
	}

	title := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.Title }),
		metadata.OGTitle,
		metadata.TwitterTitle,
		metadata.HTMLTitle,
	), embed.MaxTitleCharacters)
	description := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.Description }),
		metadata.OGDescription,
		metadata.TwitterDescription,
		metadata.Description,
	), embed.MaxDescriptionCharacters)
	providerName := firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.ProviderName }),
		metadata.SiteName,
	)
	providerURL := firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.ProviderURL }),
		siteRoot(pageURL),
	)
	authorName := truncateText(firstNonEmpty(
		valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.AuthorName }),
		metadata.AuthorName,
	), embed.MaxAuthorNameCharacters)
	authorURL := valueOrEmpty(oembedData, func(v *oEmbedResponse) string { return v.AuthorURL })
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
	if authorName != "" || authorURL != "" {
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

	if result.Title == "" && result.Description == "" && result.Provider == nil && result.Author == nil && result.Thumbnail == nil && result.Image == nil && result.Video == nil {
		return nil
	}
	return result
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

func (g *Generator) buildYouTubeOEmbedURL(rawURL string) string {
	query := url.Values{}
	query.Set("url", rawURL)
	query.Set("format", "json")
	return fmt.Sprintf("%s?%s", strings.TrimRight(g.youtubeOEmbedEndpoint, "?"), query.Encode())
}

func (g *Generator) buildYouTubeVideoURL(videoID string) string {
	return fmt.Sprintf("%s/%s", g.youtubeEmbedBaseURL, videoID)
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
