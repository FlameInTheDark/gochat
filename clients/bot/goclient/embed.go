package goclient

import (
	"fmt"
	"net/url"
	"time"
	"unicode/utf8"
)

const (
	// MaxEmbedsPerMessage is the server-side embed count limit.
	MaxEmbedsPerMessage = 10
	// MaxFieldsPerEmbed is the maximum number of fields in one embed.
	MaxFieldsPerEmbed = 25
	// MaxTotalEmbedTextCharacters is the aggregate text limit across embeds.
	MaxTotalEmbedTextCharacters = 6000
)

const (
	maxTitleCharacters       = 256
	maxDescriptionCharacters = 4096
	maxFooterTextCharacters  = 2048
	maxAuthorNameCharacters  = 256
	maxFieldNameCharacters   = 256
	maxFieldValueCharacters  = 1024
	maxColorValue            = 16777215
)

var allowedEmbedTypes = map[string]struct{}{
	"rich":    {},
	"image":   {},
	"video":   {},
	"gifv":    {},
	"article": {},
	"link":    {},
}

// Embed is a rich message embed object.
type Embed struct {
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Timestamp   *time.Time     `json:"timestamp,omitempty"`
	Color       *int           `json:"color,omitempty"`
	Footer      *EmbedFooter   `json:"footer,omitempty"`
	Image       *EmbedMedia    `json:"image,omitempty"`
	Thumbnail   *EmbedMedia    `json:"thumbnail,omitempty"`
	Video       *EmbedMedia    `json:"video,omitempty"`
	Provider    *EmbedProvider `json:"provider,omitempty"`
	Author      *EmbedAuthor   `json:"author,omitempty"`
	Fields      []EmbedField   `json:"fields,omitempty"`
}

// EmbedFooter is the footer block shown at the bottom of an embed.
type EmbedFooter struct {
	Text         string `json:"text,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedMedia describes image, thumbnail, or video media attached to an embed.
type EmbedMedia struct {
	URL                string `json:"url,omitempty"`
	ProxyURL           string `json:"proxy_url,omitempty"`
	Height             *int64 `json:"height,omitempty"`
	Width              *int64 `json:"width,omitempty"`
	ContentType        string `json:"content_type,omitempty"`
	Placeholder        string `json:"placeholder,omitempty"`
	PlaceholderVersion *int   `json:"placeholder_version,omitempty"`
	Flags              *int   `json:"flags,omitempty"`
}

// EmbedProvider identifies the site or service that produced an embed.
type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// EmbedAuthor identifies the embed author or channel.
type EmbedAuthor struct {
	Name         string `json:"name,omitempty"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedField is a structured name/value row within an embed.
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline *bool  `json:"inline,omitempty"`
}

// ValidateEmbeds validates embeds against the server's current limits.
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
	if totalText > MaxTotalEmbedTextCharacters {
		return fmt.Errorf("embeds text must not exceed %d characters in total", MaxTotalEmbedTextCharacters)
	}
	return nil
}

func validateEmbed(index int, embed Embed, totalText *int) error {
	prefix := fmt.Sprintf("embeds[%d]", index)
	if embed.Type != "" {
		if _, ok := allowedEmbedTypes[embed.Type]; !ok {
			return fmt.Errorf("%s.type must be one of rich, image, video, gifv, article, link", prefix)
		}
	}
	if err := validateLength(prefix+".title", embed.Title, maxTitleCharacters); err != nil {
		return err
	}
	if err := validateLength(prefix+".description", embed.Description, maxDescriptionCharacters); err != nil {
		return err
	}
	if err := validateURL(prefix+".url", embed.URL, true); err != nil {
		return err
	}
	if embed.Color != nil && (*embed.Color < 0 || *embed.Color > maxColorValue) {
		return fmt.Errorf("%s.color must be between 0 and %d", prefix, maxColorValue)
	}
	addText(totalText, embed.Title)
	addText(totalText, embed.Description)
	if embed.Footer != nil {
		if embed.Footer.Text == "" {
			return fmt.Errorf("%s.footer.text is required when footer is set", prefix)
		}
		if err := validateLength(prefix+".footer.text", embed.Footer.Text, maxFooterTextCharacters); err != nil {
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
		if err := validateLength(prefix+".author.name", embed.Author.Name, maxAuthorNameCharacters); err != nil {
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
		if err := validateLength(fieldPrefix+".name", field.Name, maxFieldNameCharacters); err != nil {
			return err
		}
		if err := validateLength(fieldPrefix+".value", field.Value, maxFieldValueCharacters); err != nil {
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
