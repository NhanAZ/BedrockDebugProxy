package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/buildinfo"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/proxy"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/oauth2"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "run":
		return runProxy(args[1:], stdout, stderr)
	case "verify":
		return runVerify(args[1:], stdout, stderr)
	case "version":
		_, _ = fmt.Fprintf(stdout, "bedrock-debug-proxy %s", buildinfo.Version)
		if buildinfo.Commit != "" {
			_, _ = fmt.Fprintf(stdout, " (%s)", buildinfo.Commit)
		}
		_, _ = fmt.Fprintf(stdout, " - Bedrock %s protocol %d\n", protocol.CurrentVersion, protocol.CurrentProtocol)
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown command %q.\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runProxy(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	listen := flags.String("listen", "127.0.0.1:19132", "local Bedrock address")
	upstream := flags.String("upstream", "", "destination Bedrock address")
	capturePath := flags.String("capture", "", "exact capture directory")
	captureRoot := flags.String("capture-root", "captures", "parent for generated capture directories")
	authMode := flags.String("auth", "device", "upstream authentication mode - device or none")
	allowUnauthenticated := flags.Bool("allow-unauthenticated-client", false, "disable Xbox authentication for the connecting client")
	syncEachEvent := flags.Bool("sync-each-event", true, "sync every event before forwarding continues")
	maxDecompressed := flags.Int("max-decompressed-bytes", 64<<20, "downstream decompressed batch limit")
	binaryPreview := flags.Int("decoded-binary-preview", 32, "decoded binary preview bytes")
	maxCollection := flags.Int("max-decoded-items", 0, "decoded items per collection - zero keeps all")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy run --upstream HOST:PORT [options]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "Unexpected positional arguments - %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	if *upstream == "" {
		_, _ = fmt.Fprintln(stderr, "--upstream is required.")
		return 2
	}
	var tokenSource oauth2.TokenSource
	switch *authMode {
	case "device":
		tokenSource = auth.WriterTokenSource(stderr)
	case "none":
	case "":
		_, _ = fmt.Fprintln(stderr, "--auth cannot be empty.")
		return 2
	default:
		_, _ = fmt.Fprintf(stderr, "Unsupported authentication mode %q. Use device or none.\n", *authMode)
		return 2
	}

	exactCapturePath := *capturePath
	if exactCapturePath == "" {
		var err error
		exactCapturePath, err = nextCapturePath(*captureRoot, time.Now())
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Choose capture path - %v\n", err)
			return 1
		}
	}
	recorder, err := capture.New(exactCapturePath, capture.Options{
		GeneratorName: "bedrock-debug-proxy",
		Version:       buildinfo.Version,
		Commit:        buildinfo.Commit,
		SyncEachEvent: *syncEachEvent,
		RawLayers:     []string{"bedrock_transport_payload", "bedrock_packet_payload"},
		Values: map[string]string{
			"listen_address":   *listen,
			"upstream_address": *upstream,
			"auth_mode":        *authMode,
			"protocol_id":      strconv.FormatInt(int64(protocol.CurrentProtocol), 10),
			"game_version":     protocol.CurrentVersion,
		},
		Limitations: []string{
			"Raw UDP datagrams and RakNet acknowledgement or retransmission frames are not captured",
			"This version accepts one client and records one upstream hop per process",
			"Transfer packets are recorded but automatic hop following is not implemented",
			"The upstream resource-pack-required flag is not mirrored to the downstream listener",
			"Resource pack archives and content keys are stored but decryption and extraction are not implemented",
			"Transport payload encryption and compression state is not yet classified per event",
		},
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Create capture - %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Capture directory %s\n", recorder.Root())

	runner, err := proxy.New(proxy.Config{
		ListenAddress:              *listen,
		UpstreamAddress:            *upstream,
		AllowUnauthenticatedClient: *allowUnauthenticated,
		TokenSource:                tokenSource,
		Recorder:                   recorder,
		Output:                     stdout,
		MaxDecompressedBytes:       *maxDecompressed,
		DecodedBinaryPreviewBytes:  *binaryPreview,
		MaxDecodedCollectionItems:  *maxCollection,
	})
	if err != nil {
		_ = recorder.Close("failed", err)
		_, _ = fmt.Fprintf(stderr, "Configure proxy - %v\n", err)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runErr := runner.Run(ctx)
	status := "closed"
	if runErr != nil {
		status = "failed"
	}
	closeErr := recorder.Close(status, runErr)
	if closeErr != nil {
		runErr = errors.Join(runErr, closeErr)
	}
	verification, verifyErr := capture.Verify(recorder.Root())
	if verifyErr != nil {
		runErr = errors.Join(runErr, verifyErr)
	} else if len(verification.Issues) != 0 {
		runErr = errors.Join(runErr, fmt.Errorf("capture verification found %d issue(s)", len(verification.Issues)))
		for _, issue := range verification.Issues {
			_, _ = fmt.Fprintf(stderr, "Capture issue - %s\n", issue)
		}
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stderr, "Proxy stopped with an error - %v\n", runErr)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Capture verified with %d events and %d blobs.\n", verification.Events, verification.Blobs)
	return 0
}

func runVerify(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy verify CAPTURE_DIRECTORY")
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	result, err := capture.Verify(flags.Arg(0))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Verify capture - %v\n", err)
		return 1
	}
	output := struct {
		CaptureID string   `json:"capture_id"`
		Status    string   `json:"status"`
		Events    uint64   `json:"events"`
		Blobs     uint64   `json:"blobs"`
		Issues    []string `json:"issues"`
	}{
		CaptureID: result.Manifest.CaptureID,
		Status:    result.Manifest.Status,
		Events:    result.Events,
		Blobs:     result.Blobs,
		Issues:    result.Issues,
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		_, _ = fmt.Fprintf(stderr, "Write verification result - %v\n", err)
		return 1
	}
	if len(result.Issues) != 0 {
		return 1
	}
	return 0
}

func nextCapturePath(root string, now time.Time) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("capture root is empty")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", fmt.Errorf("create capture root: %w", err)
	}
	base := "session-" + now.UTC().Format("20060102T150405Z")
	for attempt := 0; attempt < 1000; attempt++ {
		name := base
		if attempt != 0 {
			name += "-" + strconv.Itoa(attempt)
		}
		candidate := filepath.Join(root, name)
		_, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("inspect candidate %s: %w", candidate, err)
		}
	}
	return "", errors.New("capture path attempts exhausted")
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "BedrockDebugProxy")
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintln(writer, "Usage")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy run --upstream HOST:PORT [options]")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy verify CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy version")
}
