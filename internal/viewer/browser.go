package viewer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

func OpenBrowser(address string) error {
	parsed, err := url.ParseRequestURI(address)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" {
		return fmt.Errorf("refuse non-loopback viewer URL %q", address)
	}
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// #nosec G204 -- the executable is fixed and address is a validated loopback HTTP URL.
		command = exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", address)
	case "darwin":
		// #nosec G204 -- the executable is fixed and address is a validated loopback HTTP URL.
		command = exec.CommandContext(context.Background(), "open", address)
	case "linux":
		// #nosec G204 -- the executable is fixed and address is a validated loopback HTTP URL.
		command = exec.CommandContext(context.Background(), "xdg-open", address)
	default:
		return errors.New("automatic browser opening is not supported on this platform")
	}
	return command.Start()
}
