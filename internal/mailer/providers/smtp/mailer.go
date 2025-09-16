package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	smtppkg "net/smtp"

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
	// Build a minimal MIME message with HTML body, base64-encoded (RFC 2045/2047-friendly).
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s <%s>\r\n", notify.From.Name, notify.From.Email)
	fmt.Fprintf(&msg, "To: %s <%s>\r\n", notify.To.Name, notify.To.Email)
	fmt.Fprintf(&msg, "Subject: %s\r\n", notify.Subject)
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: base64\r\n\r\n")

	b64 := base64.StdEncoding.EncodeToString([]byte(notify.Html))
	wrap76(&msg, b64)

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

	if m.useTls {
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
