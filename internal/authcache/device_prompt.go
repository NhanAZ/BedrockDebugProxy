package authcache

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

const deviceAuthLoginURL = "https://login.live.com/oauth20_remoteconnect.srf?otc="

// clickableAuthWriter adapts gophertunnel's terminal device-auth prompt for
// terminals that support Ctrl+click URLs. The upstream prompt remains the
// fallback because the direct URL is an ecosystem convention rather than part
// of gophertunnel's public API.
type clickableAuthWriter struct {
	target io.Writer
}

func newClickableAuthWriter(target io.Writer) io.Writer {
	return clickableAuthWriter{target: target}
}

func (w clickableAuthWriter) Write(p []byte) (int, error) {
	text := string(p)
	if rewritten, ok := rewriteDeviceAuthPrompt(text); ok {
		text = rewritten
	}
	_, err := io.WriteString(w.target, text)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func rewriteDeviceAuthPrompt(text string) (string, bool) {
	lineEnding := ""
	if strings.HasSuffix(text, "\n") {
		lineEnding = "\n"
		text = strings.TrimSuffix(text, "\n")
	}
	if strings.HasSuffix(text, "\r") {
		lineEnding = "\r\n"
		text = strings.TrimSuffix(text, "\r")
	}

	const (
		prefix = "Authenticate at "
		marker = " using the code "
	)
	if !strings.HasPrefix(text, prefix) {
		return "", false
	}
	remainder := strings.TrimPrefix(text, prefix)
	separator := strings.Index(remainder, marker)
	if separator <= 0 {
		return "", false
	}
	verificationURI := remainder[:separator]
	code := strings.TrimSuffix(remainder[separator+len(marker):], ".")
	if code == "" || verificationURI == "" {
		return "", false
	}

	loginURL := deviceAuthLoginURL + url.QueryEscape(code)
	return fmt.Sprintf("Authenticate at %s (fallback: %s with code %s)%s", loginURL, verificationURI, code, lineEnding), true
}
