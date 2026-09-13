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

	"github.com/NhanAZ/BedrockDebugProxy/internal/analysis"
	"github.com/NhanAZ/BedrockDebugProxy/internal/artifacts"
	"github.com/NhanAZ/BedrockDebugProxy/internal/authcache"
	"github.com/NhanAZ/BedrockDebugProxy/internal/buildinfo"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capturearchive"
	"github.com/NhanAZ/BedrockDebugProxy/internal/experience"
	"github.com/NhanAZ/BedrockDebugProxy/internal/proxy"
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
	case "logout":
		return runLogout(args[1:], stdout, stderr)
	case "verify":
		return runVerify(args[1:], stdout, stderr)
	case "inspect":
		return runInspect(args[1:], stdout, stderr)
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	case "explain":
		return runExplain(args[1:], stdout, stderr)
	case "export":
		return runExport(args[1:], stdout, stderr)
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

func runLogout(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("logout", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy logout")
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "Unexpected positional arguments - %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	cachePath, err := authcache.DefaultPath()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Locate authentication cache - %v\n", err)
		return 1
	}
	removed, err := authcache.Remove(cachePath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Logout - %v\n", err)
		return 1
	}
	if removed {
		_, _ = fmt.Fprintf(stdout, "Removed cached Microsoft authentication from %s.\n", cachePath)
	} else {
		_, _ = fmt.Fprintf(stdout, "No cached Microsoft authentication found at %s.\n", cachePath)
	}
	return 0
}

func runProxy(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	listen := flags.String("listen", "127.0.0.1:19132", "local Bedrock address")
	upstream := flags.String("upstream", "", "destination HOST:PORT or experience:NAME")
	capturePath := flags.String("capture", "", "exact capture directory")
	captureRoot := flags.String("capture-root", "captures", "parent for generated capture directories")
	authMode := flags.String("auth", "device", "upstream authentication mode - device (cached) or none")
	allowUnauthenticated := flags.Bool("allow-unauthenticated-client", false, "disable Xbox authentication for the connecting client")
	followTransfers := flags.Bool("follow-transfers", false, "rewrite server transfers to the local listener and follow subsequent hops")
	syncEachEvent := flags.Bool("sync-each-event", false, "durably sync every blob and event before forwarding continues")
	maxDecompressed := flags.Int("max-decompressed-bytes", 64<<20, "downstream decompressed batch limit")
	binaryPreview := flags.Int("decoded-binary-preview", 32, "decoded binary preview bytes")
	maxCollection := flags.Int("max-decoded-items", 0, "decoded items per collection - zero keeps all")
	decryptResourcePacks := flags.Bool("decrypt-resource-packs", false, "derive decrypted resource pack artifacts when a supported content key is available")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy run --upstream HOST:PORT|experience:NAME [options]")
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
		cachePath, err := authcache.DefaultPath()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Configure authentication - %v\n", err)
			return 1
		}
		tokenSource, err = authcache.New(cachePath, stderr)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Configure authentication - %v\n", err)
			return 1
		}
	case "none":
	case "":
		_, _ = fmt.Fprintln(stderr, "--auth cannot be empty.")
		return 2
	default:
		_, _ = fmt.Fprintf(stderr, "Unsupported authentication mode %q. Use device or none.\n", *authMode)
		return 2
	}
	if experience.IsSelector(*upstream) && tokenSource == nil {
		_, _ = fmt.Fprintln(stderr, "Experience targets require --auth device.")
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	target, err := experience.Resolve(ctx, *upstream, tokenSource, nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Resolve upstream - %v\n", err)
		return 1
	}
	if experience.IsSelector(*upstream) {
		_, _ = fmt.Fprintf(stdout, "Resolved experience %q using %s.\n", target.Name, target.Transport)
	}

	exactCapturePath := *capturePath
	if exactCapturePath == "" {
		var err error
		exactCapturePath, err = nextCapturePath(*captureRoot, time.Now())
		if err != nil {
			_ = target.Close()
			_, _ = fmt.Fprintf(stderr, "Choose capture path - %v\n", err)
			return 1
		}
	}
	limitations := []string{
		"Raw UDP datagrams and RakNet acknowledgement or retransmission frames are not captured",
		"Downstream clients must accept offered resource packs; the current adapter cannot mirror an upstream optional-pack policy",
		"Transport payload encryption and compression state is not yet classified per event",
	}
	if *followTransfers {
		limitations = append(limitations, "Transfer following is opt-in; server Transfer packets are rewritten to the local listener and each subsequent hop is captured in the same session")
	} else {
		limitations = append(limitations, "Transfer packets are recorded but automatic hop following is disabled")
	}
	if *decryptResourcePacks {
		limitations = append(limitations, "Resource pack decryption supports only the documented AES-256-CFB8 contents format; unsupported variants are retained with a decrypt error")
	} else {
		limitations = append(limitations, "Encrypted resource pack archives and content keys are stored; decryption was not requested")
	}
	recorder, err := capture.New(exactCapturePath, capture.Options{
		GeneratorName: "bedrock-debug-proxy",
		Version:       buildinfo.Version,
		Commit:        buildinfo.Commit,
		SyncEachEvent: *syncEachEvent,
		RawLayers:     []string{"bedrock_transport_payload", "bedrock_packet_payload"},
		Values: map[string]string{
			"listen_address":               *listen,
			"upstream_selector":            target.Selector,
			"upstream_address":             target.Address,
			"upstream_transport":           target.Transport,
			"upstream_name":                target.Name,
			"upstream_experience_id":       target.ExperienceID,
			"auth_mode":                    *authMode,
			"allow_unauthenticated_client": strconv.FormatBool(*allowUnauthenticated),
			"follow_transfers":             strconv.FormatBool(*followTransfers),
			"protocol_id":                  strconv.FormatInt(int64(protocol.CurrentProtocol), 10),
			"game_version":                 protocol.CurrentVersion,
			"decrypt_resource_packs":       strconv.FormatBool(*decryptResourcePacks),
			"automatic_artifacts":          artifacts.Schema,
		},
		Limitations: limitations,
	})
	if err != nil {
		_ = target.Close()
		_, _ = fmt.Fprintf(stderr, "Create capture - %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Capture directory %s\n", recorder.Root())

	runner, err := proxy.New(proxy.Config{
		ListenAddress:              *listen,
		UpstreamAddress:            target.Address,
		UpstreamNetwork:            target.Network,
		AllowUnauthenticatedClient: *allowUnauthenticated,
		FollowTransfers:            *followTransfers,
		TokenSource:                target.TokenSource,
		Recorder:                   recorder,
		Output:                     stdout,
		MaxDecompressedBytes:       *maxDecompressed,
		DecodedBinaryPreviewBytes:  *binaryPreview,
		MaxDecodedCollectionItems:  *maxCollection,
		DecryptResourcePacks:       *decryptResourcePacks,
	})
	if err != nil {
		err = errors.Join(err, target.Close())
		_ = recorder.Close("failed", err)
		_, _ = fmt.Fprintf(stderr, "Configure proxy - %v\n", err)
		return 2
	}
	views, err := artifacts.Start(recorder.Root(), protocol.CurrentProtocol, func() {
		_, _ = fmt.Fprintln(stderr, "Automatic artifact output encountered an error. The original capture is unchanged. See artifacts/status.json and artifacts/errors.jsonl.")
		_ = recorder.AddLimitation("Automatic artifact output encountered an error; inspect artifacts/status.json and artifacts/errors.jsonl. Original capture evidence is retained")
	})
	if err != nil {
		_ = target.Close()
		_ = recorder.Close("failed", err)
		_, _ = fmt.Fprintf(stderr, "Create automatic artifact folders - %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Automatic artifact folders %s\n", filepath.Join(recorder.Root(), "artifacts"))
	runErr := runner.Run(ctx)
	runErr = errors.Join(runErr, target.Close())
	status := "closed"
	if runErr != nil {
		status = "failed"
	}
	closeErr := recorder.Close(status, runErr)
	if closeErr != nil {
		runErr = errors.Join(runErr, closeErr)
	}
	_, _ = fmt.Fprintln(stdout, "Finishing automatic artifact folders...")
	artifactStatus, artifactErr := views.Finish()
	runErr = errors.Join(runErr, artifactErr)
	_, _ = fmt.Fprintf(stdout, "Automatic artifacts %s through event %d.\n", artifactStatus.State, artifactStatus.LastSequence)
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

func runInspect(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.SetOutput(stderr)
	kind := flags.String("kind", "", "exact event kind")
	packetName := flags.String("packet", "", "exact decoded packet name")
	direction := flags.String("direction", "", "exact event direction")
	channel := flags.String("channel", "", "exact event channel")
	fromSequence := flags.Uint64("from-sequence", 0, "minimum event sequence")
	limit := flags.Uint64("limit", 0, "maximum matching events - zero keeps all")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy inspect [filters] CAPTURE_DIRECTORY")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if !validDirectionFilter(*direction) {
		_, _ = fmt.Fprintf(stderr, "Unsupported direction %q.\n", *direction)
		return 2
	}
	verification, err := capture.Verify(flags.Arg(0))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Inspect capture - %v\n", err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	var matched uint64
	err = capture.ScanEvents(flags.Arg(0), func(event capture.Event) error {
		if event.Sequence < *fromSequence || *kind != "" && event.Kind != *kind ||
			*direction != "" && string(event.Direction) != *direction || *channel != "" && event.Channel != *channel ||
			*packetName != "" && (event.Packet == nil || event.Packet.Name != *packetName) {
			return nil
		}
		if *limit != 0 && matched >= *limit {
			return nil
		}
		matched++
		return encoder.Encode(event)
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Inspect capture - %v\n", err)
		return 1
	}
	if len(verification.Issues) != 0 {
		for _, issue := range verification.Issues {
			_, _ = fmt.Fprintf(stderr, "Capture issue - %s\n", issue)
		}
		return 1
	}
	return 0
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy analyze CAPTURE_DIRECTORY")
		return 2
	}
	summary, err := analysis.Analyze(args[0])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Analyze capture - %v\n", err)
		return 1
	}
	if err := writeJSON(stdout, summary); err != nil {
		_, _ = fmt.Fprintf(stderr, "Write analysis - %v\n", err)
		return 1
	}
	if len(summary.VerificationIssues) != 0 {
		return 1
	}
	return 0
}

func runExplain(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy explain CAPTURE_DIRECTORY")
		return 2
	}
	summary, err := analysis.Analyze(args[0])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Explain capture - %v\n", err)
		return 1
	}
	if _, err := io.WriteString(stdout, analysis.Explain(summary)); err != nil {
		_, _ = fmt.Fprintf(stderr, "Write explanation - %v\n", err)
		return 1
	}
	if len(summary.VerificationIssues) != 0 {
		return 1
	}
	return 0
}

func runExport(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy export CAPTURE_DIRECTORY OUTPUT.bdpcap")
		return 2
	}
	result, err := capturearchive.Export(args[0], args[1])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Export capture - %v\n", err)
		return 1
	}
	if result.ContainsDecryptedResourcePacks {
		_, _ = fmt.Fprintln(stderr, "Warning - this local export contains decrypted resource-pack assets and keys whose ownership and redistribution rights are not granted by the BedrockDebugProxy license.")
	}
	if err := writeJSON(stdout, result); err != nil {
		_, _ = fmt.Fprintf(stderr, "Write export result - %v\n", err)
		return 1
	}
	return 0
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func validDirectionFilter(value string) bool {
	switch capture.Direction(value) {
	case "", capture.DirectionClientToServer, capture.DirectionServerToClient, capture.DirectionInternal, capture.DirectionUnknown:
		return true
	default:
		return false
	}
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
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy run --upstream HOST:PORT|experience:NAME [options]")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy logout")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy verify CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy inspect [filters] CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy analyze CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy explain CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy export CAPTURE_DIRECTORY OUTPUT.bdpcap")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy version")
}
