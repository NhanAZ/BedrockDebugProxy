package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft"
	"golang.org/x/oauth2"
)

type Config struct {
	ListenAddress              string
	UpstreamAddress            string
	UpstreamNetwork            minecraft.Network
	AllowUnauthenticatedClient bool
	FollowTransfers            bool
	TokenSource                oauth2.TokenSource
	Recorder                   *capture.Recorder
	Output                     io.Writer
	MaxDecompressedBytes       int
	DecodedBinaryPreviewBytes  int
	MaxDecodedCollectionItems  int
	DecryptResourcePacks       bool
}

func (c *Config) normalize() error {
	if c.Recorder == nil {
		return errors.New("capture recorder is required")
	}
	if c.ListenAddress == "" {
		c.ListenAddress = "127.0.0.1:19132"
	}
	if c.UpstreamAddress == "" {
		return errors.New("upstream address is required")
	}
	if err := validateAddress(c.ListenAddress); err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	if c.UpstreamNetwork == nil {
		if err := validateAddress(c.UpstreamAddress); err != nil {
			return fmt.Errorf("invalid upstream address: %w", err)
		}
	} else if strings.TrimSpace(c.UpstreamAddress) == "" {
		return errors.New("invalid upstream address: address is empty")
	} else if strings.ContainsAny(c.UpstreamAddress, "\r\n") {
		return errors.New("invalid upstream address: address contains a line break")
	}
	if c.Output == nil {
		c.Output = io.Discard
	}
	if c.MaxDecompressedBytes == 0 {
		c.MaxDecompressedBytes = 64 << 20
	}
	if c.MaxDecompressedBytes < 0 {
		return errors.New("maximum decompressed bytes cannot be negative")
	}
	if c.DecodedBinaryPreviewBytes == 0 {
		c.DecodedBinaryPreviewBytes = 32
	}
	if c.DecodedBinaryPreviewBytes < 0 {
		return errors.New("decoded binary preview bytes cannot be negative")
	}
	if c.MaxDecodedCollectionItems < 0 {
		return errors.New("maximum decoded collection items cannot be negative")
	}
	return nil
}

func validateAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if strings.TrimSpace(port) == "" {
		return errors.New("port is empty")
	}
	if strings.ContainsAny(host, "\r\n") {
		return errors.New("host contains a line break")
	}
	return nil
}
