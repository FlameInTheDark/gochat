package smtp

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	smtppkg "net/smtp"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/mailer"
)

// SmtpMailer sends emails via a plain SMTP server.
// It mirrors the Send(ctx, notify) signature used by the SendPulse-based mailer.
//
// addr should be in the form "host:port" (e.g., "smtp.gmail.com:587" or
// "smtp.mailgun.org:465"). Port 465 will use implicit TLS; otherwise we'll try STARTTLS.
// username/password are used for AUTH PLAIN/LOGIN if the server supports AUTH.
//
// Note: net/smtp is frozen but stable. For richer needs consider third-party libs.
// This implementation is context-aware for dialing and supports TLS and STARTTLS.
type SmtpMailer struct {
	host     string
	port     int
	username string
	password string
	useTls   bool
}

func New(host string, port int, username, password string, useTls bool) *SmtpMailer {
	return &SmtpMailer{host: host, port: port, username: username, password: password, useTls: useTls}
}

func (m *SmtpMailer) Send(ctx context.Context, notify mailer.MailNotification) error {
	// Build a MIME message with Date and Message-ID, and multipart/alternative body.
	var msg bytes.Buffer
	fromName := strings.TrimSpace(notify.From.Name)
	toName := strings.TrimSpace(notify.To.Name)
	fromEmail := strings.TrimSpace(notify.From.Email)
	toEmail := strings.TrimSpace(notify.To.Email)

	// Headers
	fmt.Fprintf(&msg, "From: %s <%s>\r\n", fromName, fromEmail)
	fmt.Fprintf(&msg, "To: %s <%s>\r\n", toName, toEmail)
	fmt.Fprintf(&msg, "Subject: %s\r\n", notify.Subject)
	fmt.Fprintf(&msg, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&msg, "Message-ID: %s\r\n", genMessageID(domainPart(fromEmail, m.host)))
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")

	// multipart/alternative boundary
	boundary := genBoundary()
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)

	// text/plain part
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: base64\r\n\r\n")
	textAlt := htmlToText(notify.Html)
	wrap76(&msg, base64.StdEncoding.EncodeToString([]byte(textAlt)))

	// text/html part
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: base64\r\n\r\n")
	wrap76(&msg, base64.StdEncoding.EncodeToString([]byte(notify.Html)))

	// closing boundary
	fmt.Fprintf(&msg, "--%s--\r\n", boundary)

	var (
		conn net.Conn
		err  error
	)
	if m.useTls {
		conn, err = tls.DialWithDialer(
			&net.Dialer{},
			"tcp",
			fmt.Sprintf("%s:%d", m.host, m.port),
			&tls.Config{ServerName: m.host},
		)
		if err != nil {
			return err
		}
	} else {
		// Plain TCP first; we'll upgrade via STARTTLS if the server supports it
		dialer := &net.Dialer{}
		conn, err = dialer.DialContext(
			ctx,
			"tcp",
			fmt.Sprintf("%s:%d", m.host, m.port),
		)
		if err != nil {
			return err
		}
	}
	defer conn.Close()

	c, err := smtppkg.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer c.Close()

	// If we connected without implicit TLS, try STARTTLS when available
	if !m.useTls {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
				return err
			}
		}
	}

	// AUTH when supported
	if m.username != "" || m.password != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			auth := smtppkg.PlainAuth("", m.username, m.password, m.host)
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := c.Mail(notify.From.Email); err != nil {
		return err
	}
	if err := c.Rcpt(notify.To.Email); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(msg.Bytes()); err != nil {
		_ = wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return c.Quit()
}

// wrap76 writes the base64 body split into 76-character lines as recommended by RFC 2045.
func wrap76(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); i += 76 {
		end := i + 76
		if end > len(s) {
			end = len(s)
		}
		buf.WriteString(s[i:end])
		buf.WriteString("\r\n")
	}
}

// genBoundary returns a random-ish MIME boundary token.
func genBoundary() string {
	b := make([]byte, 12)
	if _, err := crand.Read(b); err != nil {
		// fallback to time-based when entropy unavailable
		return fmt.Sprintf("gc-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("gc-%s", hex.EncodeToString(b))
}

// genMessageID creates a RFC 2822-style Message-ID using time and randomness.
func genMessageID(domain string) string {
	b := make([]byte, 12)
	if _, err := crand.Read(b); err != nil {
		return fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), domain)
	}
	return fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), hex.EncodeToString(b), domain)
}

// domainPart extracts the domain from an email address, falling back to fallback.
func domainPart(email, fallback string) string {
	if at := strings.LastIndex(email, "@"); at != -1 && at+1 < len(email) {
		return email[at+1:]
	}
	return fallback
}

// htmlToText produces a minimal text alternative from HTML.
func htmlToText(html string) string {
	if html == "" {
		return ""
	}
	// Basic block-level replacements for readability
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"</p>", "\n\n",
		"</div>", "\n\n",
		"</li>", "\n",
		"</h1>", "\n\n",
		"</h2>", "\n\n",
		"</h3>", "\n\n",
		"&nbsp;", " ",
	)
	s := replacer.Replace(html)
	// Strip tags
	var b strings.Builder
	inTag := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteByte(c)
			}
		}
	}
	// Collapse excessive blank lines
	out := b.String()
	out = strings.ReplaceAll(out, "\r", "")
	lines := strings.Split(out, "\n")
	var compact []string
	lastBlank := false
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if trimmed == "" {
			if !lastBlank {
				compact = append(compact, "")
				lastBlank = true
			}
			continue
		}
		compact = append(compact, trimmed)
		lastBlank = false
	}
	return strings.TrimSpace(strings.Join(compact, "\n"))
}
