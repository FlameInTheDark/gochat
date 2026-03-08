package embed

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxEmbedsPerMessage      = 10
	MaxFieldsPerEmbed        = 25
	MaxTotalTextCharacters   = 6000
	MaxTitleCharacters       = 256
	MaxDescriptionCharacters = 4096
	MaxFooterTextCharacters  = 2048
	MaxAuthorNameCharacters  = 256
	MaxFieldNameCharacters   = 256
	MaxFieldValueCharacters  = 1024
	MaxColorValue            = 16777215
)

var allowedEmbedTypes = map[string]struct{}{
	"rich":    {},
	"image":   {},
	"video":   {},
	"gifv":    {},
	"article": {},
	"link":    {},
}

// Embed is a Discord-like message embed object.
type Embed struct {
	Title       string         `json:"title,omitempty" example:"GoChat 1.0"`                                     // Embed title.
	Type        string         `json:"type,omitempty" example:"rich" enums:"rich,image,video,gifv,article,link"` // Embed type.
	Description string         `json:"description,omitempty" example:"Embed support is live."`                   // Main embed description.
	URL         string         `json:"url,omitempty" example:"https://example.com/release"`                      // Canonical URL opened when the embed title is clicked.
	Timestamp   *time.Time     `json:"timestamp,omitempty" swaggertype:"string" format:"date-time"`              // Optional ISO timestamp shown by the client.
	Color       *int           `json:"color,omitempty" example:"65280"`                                          // Decimal RGB color value.
	Footer      *EmbedFooter   `json:"footer,omitempty"`                                                         // Optional footer block.
	Image       *EmbedMedia    `json:"image,omitempty"`                                                          // Full-size image block.
	Thumbnail   *EmbedMedia    `json:"thumbnail,omitempty"`                                                      // Thumbnail image block.
	Video       *EmbedMedia    `json:"video,omitempty"`                                                          // Embedded video metadata.
	Provider    *EmbedProvider `json:"provider,omitempty"`                                                       // Content provider metadata.
	Author      *EmbedAuthor   `json:"author,omitempty"`                                                         // Embed author metadata.
	Fields      []EmbedField   `json:"fields,omitempty"`                                                         // Up to 25 structured fields.
}

// EmbedFooter is the footer block shown at the bottom of an embed.
type EmbedFooter struct {
	Text         string `json:"text,omitempty" example:"Documentation"`                              // Footer text.
	IconURL      string `json:"icon_url,omitempty" example:"https://example.com/icon.png"`           // Footer icon URL.
	ProxyIconURL string `json:"proxy_icon_url,omitempty" example:"https://cdn.example.com/icon.png"` // Optional proxied footer icon URL.
}

// EmbedMedia describes image, thumbnail, or video media attached to an embed.
type EmbedMedia struct {
	URL                string `json:"url,omitempty" example:"https://example.com/preview.png"`           // Media URL.
	ProxyURL           string `json:"proxy_url,omitempty" example:"https://cdn.example.com/preview.png"` // Optional proxied media URL.
	Height             *int64 `json:"height,omitempty" example:"720"`                                    // Media height in pixels.
	Width              *int64 `json:"width,omitempty" example:"1280"`                                    // Media width in pixels.
	ContentType        string `json:"content_type,omitempty" example:"image/png"`                        // Media MIME type when known.
	Placeholder        string `json:"placeholder,omitempty" example:"IPgNDIQbRUaQiJcwuYlwmyQ5Bg=="`      // Encoded placeholder used by some generated embeds.
	PlaceholderVersion *int   `json:"placeholder_version,omitempty" example:"1"`                         // Placeholder format version.
	Flags              *int   `json:"flags,omitempty" example:"0"`                                       // Media-specific flags from the source provider.
}

// EmbedProvider identifies the site or service that produced an embed.
type EmbedProvider struct {
	Name string `json:"name,omitempty" example:"YouTube"`                // Provider name.
	URL  string `json:"url,omitempty" example:"https://www.youtube.com"` // Provider URL.
}

// EmbedAuthor identifies the embed author or channel.
type EmbedAuthor struct {
	Name         string `json:"name,omitempty" example:"The PrimeTime"`                                           // Author display name.
	URL          string `json:"url,omitempty" example:"https://www.youtube.com/channel/UCUyeluBRhGPCW4rPe_UvBZQ"` // Author URL.
	IconURL      string `json:"icon_url,omitempty" example:"https://example.com/author.png"`                      // Author icon URL.
	ProxyIconURL string `json:"proxy_icon_url,omitempty" example:"https://cdn.example.com/author.png"`            // Optional proxied author icon URL.
}

// EmbedField is a structured name/value row within an embed.
type EmbedField struct {
	Name   string `json:"name" example:"Status"`           // Field label.
	Value  string `json:"value" example:"Stable"`          // Field value.
	Inline *bool  `json:"inline,omitempty" example:"true"` // Whether the field should be rendered inline.
}

func MergeEmbeds(groups ...[]Embed) []Embed {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	if total == 0 {
		return nil
	}

	merged := make([]Embed, 0, total)
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		remaining := MaxEmbedsPerMessage - len(merged)
		if remaining <= 0 {
			break
		}
		if len(group) > remaining {
			merged = append(merged, append([]Embed(nil), group[:remaining]...)...)
			break
		}
		merged = append(merged, append([]Embed(nil), group...)...)
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func ValidateEmbeds(embeds []Embed) error {
	if len(embeds) > MaxEmbedsPerMessage {
		return fmt.Errorf("embeds must not contain more than %d items", MaxEmbedsPerMessage)
	}

	totalText := 0
	for index, embed := range embeds {
		if err := validateEmbed(index, embed, &totalText); err != nil {
			return err
		}
	}

	if totalText > MaxTotalTextCharacters {
		return fmt.Errorf("embeds text must not exceed %d characters in total", MaxTotalTextCharacters)
	}

	return nil
}

func MarshalEmbeds(embeds []Embed) (string, error) {
	if err := ValidateEmbeds(embeds); err != nil {
		return "", err
	}

	if len(embeds) == 0 {
		return "[]", nil
	}

	data, err := json.Marshal(embeds)
	if err != nil {
		return "", fmt.Errorf("unable to encode embeds: %w", err)
	}

	return string(data), nil
}

func ParseEmbeds(raw *string) ([]Embed, error) {
	if raw == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return nil, nil
	}

	var embeds []Embed
	if err := json.Unmarshal([]byte(trimmed), &embeds); err != nil {
		return nil, fmt.Errorf("unable to decode embeds: %w", err)
	}
	if err := ValidateEmbeds(embeds); err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, nil
	}

	return embeds, nil
}

func ParseMergedEmbeds(manualRaw, autoRaw *string, suppressGenerated bool) ([]Embed, error) {
	manual, err := ParseEmbeds(manualRaw)
	if err != nil {
		return nil, err
	}
	if suppressGenerated {
		return manual, nil
	}

	generated, err := ParseEmbeds(autoRaw)
	if err != nil {
		return nil, err
	}

	return MergeEmbeds(manual, generated), nil
}
func validateEmbed(index int, embed Embed, totalText *int) error {
	prefix := fmt.Sprintf("embeds[%d]", index)

	if embed.Type != "" {
		if _, ok := allowedEmbedTypes[embed.Type]; !ok {
			return fmt.Errorf("%s.type must be one of rich, image, video, gifv, article, link", prefix)
		}
	}
	if err := validateLength(prefix+".title", embed.Title, MaxTitleCharacters); err != nil {
		return err
	}
	if err := validateLength(prefix+".description", embed.Description, MaxDescriptionCharacters); err != nil {
		return err
	}
	if err := validateURL(prefix+".url", embed.URL, true); err != nil {
		return err
	}
	if embed.Color != nil && (*embed.Color < 0 || *embed.Color > MaxColorValue) {
		return fmt.Errorf("%s.color must be between 0 and %d", prefix, MaxColorValue)
	}

	addText(totalText, embed.Title)
	addText(totalText, embed.Description)

	if embed.Footer != nil {
		if embed.Footer.Text == "" {
			return fmt.Errorf("%s.footer.text is required when footer is set", prefix)
		}
		if err := validateLength(prefix+".footer.text", embed.Footer.Text, MaxFooterTextCharacters); err != nil {
			return err
		}
		if err := validateURL(prefix+".footer.icon_url", embed.Footer.IconURL, false); err != nil {
			return err
		}
		if err := validateURL(prefix+".footer.proxy_icon_url", embed.Footer.ProxyIconURL, false); err != nil {
			return err
		}
		addText(totalText, embed.Footer.Text)
	}

	if embed.Author != nil {
		if embed.Author.Name == "" {
			return fmt.Errorf("%s.author.name is required when author is set", prefix)
		}
		if err := validateLength(prefix+".author.name", embed.Author.Name, MaxAuthorNameCharacters); err != nil {
			return err
		}
		if err := validateURL(prefix+".author.url", embed.Author.URL, true); err != nil {
			return err
		}
		if err := validateURL(prefix+".author.icon_url", embed.Author.IconURL, false); err != nil {
			return err
		}
		if err := validateURL(prefix+".author.proxy_icon_url", embed.Author.ProxyIconURL, false); err != nil {
			return err
		}
		addText(totalText, embed.Author.Name)
	}

	if err := validateMedia(prefix+".image", embed.Image); err != nil {
		return err
	}
	if err := validateMedia(prefix+".thumbnail", embed.Thumbnail); err != nil {
		return err
	}
	if err := validateMedia(prefix+".video", embed.Video); err != nil {
		return err
	}

	if embed.Provider != nil {
		if err := validateURL(prefix+".provider.url", embed.Provider.URL, true); err != nil {
			return err
		}
	}

	if len(embed.Fields) > MaxFieldsPerEmbed {
		return fmt.Errorf("%s.fields must not contain more than %d items", prefix, MaxFieldsPerEmbed)
	}

	for fieldIndex, field := range embed.Fields {
		fieldPrefix := fmt.Sprintf("%s.fields[%d]", prefix, fieldIndex)
		if field.Name == "" {
			return fmt.Errorf("%s.name is required", fieldPrefix)
		}
		if field.Value == "" {
			return fmt.Errorf("%s.value is required", fieldPrefix)
		}
		if err := validateLength(fieldPrefix+".name", field.Name, MaxFieldNameCharacters); err != nil {
			return err
		}
		if err := validateLength(fieldPrefix+".value", field.Value, MaxFieldValueCharacters); err != nil {
			return err
		}
		addText(totalText, field.Name)
		addText(totalText, field.Value)
	}

	return nil
}

func validateMedia(field string, media *EmbedMedia) error {
	if media == nil {
		return nil
	}

	if err := validateURL(field+".url", media.URL, true); err != nil {
		return err
	}
	if err := validateURL(field+".proxy_url", media.ProxyURL, false); err != nil {
		return err
	}
	if media.Height != nil && *media.Height < 0 {
		return fmt.Errorf("%s.height must be non-negative", field)
	}
	if media.Width != nil && *media.Width < 0 {
		return fmt.Errorf("%s.width must be non-negative", field)
	}
	if media.PlaceholderVersion != nil && *media.PlaceholderVersion < 0 {
		return fmt.Errorf("%s.placeholder_version must be non-negative", field)
	}
	if media.Flags != nil && *media.Flags < 0 {
		return fmt.Errorf("%s.flags must be non-negative", field)
	}

	return nil
}

func validateLength(field, value string, limit int) error {
	if utf8.RuneCountInString(value) > limit {
		return fmt.Errorf("%s must not exceed %d characters", field, limit)
	}

	return nil
}

func addText(total *int, value string) {
	*total += utf8.RuneCountInString(value)
}

func validateURL(field, value string, allowAttachment bool) error {
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL", field)
	}

	switch parsed.Scheme {
	case "http", "https":
		if parsed.Host == "" {
			return fmt.Errorf("%s must be a valid URL", field)
		}
	case "attachment":
		if !allowAttachment || (parsed.Host == "" && parsed.Path == "") {
			return fmt.Errorf("%s must be a valid URL", field)
		}
	default:
		return fmt.Errorf("%s must use http, https, or attachment scheme", field)
	}

	return nil
}
