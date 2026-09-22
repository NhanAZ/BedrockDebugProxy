package main

import (
	"bufio"
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
	"github.com/NhanAZ/BedrockDebugProxy/internal/viewer"
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
	case "login":
		return runLogin(args[1:], stdout, stderr)
	case "accounts":
		return runAccounts(args[1:], stdout, stderr)
	case "use":
		return runUseAccount(args[1:], stdout, stderr)
	case "logout":
		return runLogout(args[1:], stdout, stderr)
	case "verify":
		return runVerify(args[1:], stdout, stderr)
	case "inspect":
		return runInspect(args[1:], stdout, stderr)
	case "view":
		return runView(args[1:], stdout, stderr)
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
	if flags.NArg() > 1 {
		_, _ = fmt.Fprintf(stderr, "Unexpected positional arguments - %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	store, err := authcache.NewAccountStore()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Locate authentication accounts - %v\n", err)
		return 1
	}
	accounts, err := store.List()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "List authentication accounts - %v\n", err)
		return 1
	}
	if len(accounts) == 0 {
		_, _ = fmt.Fprintln(stdout, "No cached Microsoft authentication found.")
		return 0
	}
	name := flags.Arg(0)
	if name == "" {
		defaultAccount, defaultErr := store.Default()
		if defaultErr != nil {
			_, _ = fmt.Fprintf(stderr, "Select default account - %v\n", defaultErr)
			return 1
		}
		name = defaultAccount.Name
	}
	removed, replacement, err := store.Remove(name)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Logout - %v\n", err)
		return 1
	}
	if removed {
		_, _ = fmt.Fprintf(stdout, "Logged out of account %q.\n", name)
	} else {
		_, _ = fmt.Fprintf(stdout, "Account %q had no cached Microsoft authentication.\n", name)
	}
	if replacement != "" {
		_, _ = fmt.Fprintf(stdout, "Default account is now %q.\n", replacement)
	}
	return 0
}

func runLogin(args []string, stdout, stderr io.Writer) int {
	return runLoginWithInput(args, os.Stdin, stdout, stderr)
}

func runLoginWithInput(args []string, input io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("login", flag.ContinueOnError)
	flags.SetOutput(stderr)
	maxAgeText := flags.String("max-age", "", "optional local cache age such as 7d or 168h")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy login [--max-age 7d] [ACCOUNT_NAME]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(reorderLoginFlags(args)); err != nil {
		return 2
	}
	if flags.NArg() > 1 {
		_, _ = fmt.Fprintf(stderr, "Unexpected positional arguments - %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	store, err := authcache.NewAccountStore()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Locate authentication accounts - %v\n", err)
		return 1
	}
	accounts, err := store.List()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "List authentication accounts - %v\n", err)
		return 1
	}
	interactive := isInteractiveInput(input)
	var promptReader *bufio.Reader
	if interactive {
		promptReader = bufio.NewReader(input)
	}
	name := strings.TrimSpace(flags.Arg(0))
	if name == "" && promptReader != nil {
		name, err = promptAccountName(promptReader, stdout, len(accounts) == 0)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Read account name - %v\n", err)
			return 1
		}
	}
	if name == "" {
		if len(accounts) == 0 {
			_, _ = fmt.Fprintln(stderr, "An account name is required for the first login. Run login ACCOUNT_NAME.")
			return 2
		}
		defaultAccount, defaultErr := store.Default()
		if defaultErr != nil {
			_, _ = fmt.Fprintf(stderr, "Select default account - %v\n", defaultErr)
			return 1
		}
		name = defaultAccount.Name
	}
	_, tokenPath, err := store.TokenPath(name)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Validate account name - %v\n", err)
		return 2
	}
	var cacheInput io.Reader = input
	if promptReader != nil {
		cacheInput = promptReader
	}
	maxAge, setMaxAge, err := loginCachePolicy(*maxAgeText, cacheInput, stdout, interactive)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Parse local cache age - %v\n", err)
		return 2
	}
	tokenSource, err := authcache.NewWithOptions(tokenPath, stderr, authcache.Options{ForceDevice: true})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Configure authentication - %v\n", err)
		return 1
	}
	if _, err := tokenSource.Token(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Authenticate account %q - %v\n", name, err)
		return 1
	}
	if err := store.Upsert(name, maxAge, setMaxAge); err != nil {
		_, _ = fmt.Fprintf(stderr, "Save account %q - %v\n", name, err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Account %q authenticated.\n", name)
	return 0
}

func promptAccountName(input *bufio.Reader, output io.Writer, required bool) (string, error) {
	prompt := "Microsoft account username or local label (Enter for default): "
	if required {
		prompt = "Microsoft account username or local label (required): "
	}
	if _, err := fmt.Fprint(output, prompt); err != nil {
		return "", err
	}
	line, err := input.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	name := strings.TrimSpace(line)
	if required && name == "" {
		return "", errors.New("an account name is required for the first login")
	}
	return name, nil
}

func loginCachePolicy(flagText string, input io.Reader, output io.Writer, interactive bool) (time.Duration, bool, error) {
	if strings.TrimSpace(flagText) != "" {
		maxAge, err := authcache.ParseMaxCacheAge(flagText)
		return maxAge, true, err
	}
	if !interactive {
		return 0, true, nil
	}
	_, _ = fmt.Fprint(output, "Local cache age (Enter for no limit, for example 7d): ")
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, false, err
	}
	maxAge, parseErr := authcache.ParseMaxCacheAge(strings.TrimSpace(line))
	return maxAge, true, parseErr
}

func isInteractiveInput(input io.Reader) bool {
	file, ok := input.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// reorderLoginFlags keeps the short login form convenient when an operator
// writes "login ACCOUNT --max-age 7d". The standard flag package stops at the
// first positional argument, so move the one supported option ahead of it.
func reorderLoginFlags(args []string) []string {
	options := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "--max-age" && index+1 < len(args) {
			options = append(options, argument, args[index+1])
			index++
			continue
		}
		if strings.HasPrefix(argument, "--max-age=") {
			options = append(options, argument)
			continue
		}
		positionals = append(positionals, argument)
	}
	return append(options, positionals...)
}

func runAccounts(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("accounts", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy accounts")
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "Unexpected positional arguments - %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	store, err := authcache.NewAccountStore()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Locate authentication accounts - %v\n", err)
		return 1
	}
	accounts, err := store.List()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "List authentication accounts - %v\n", err)
		return 1
	}
	if len(accounts) == 0 {
		_, _ = fmt.Fprintln(stdout, "No accounts configured. Run bedrock-debug-proxy login.")
		return 0
	}
	_, _ = fmt.Fprintln(stdout, "Accounts:")
	for _, account := range accounts {
		marker := " "
		if account.Default {
			marker = "*"
		}
		line := fmt.Sprintf("%s %s - %s", marker, account.Name, account.CacheState)
		if !account.CacheExpires.IsZero() {
			line += ", local cache until " + account.CacheExpires.Local().Format(time.RFC3339)
		}
		if !account.AccessExpires.IsZero() {
			line += ", access token until " + account.AccessExpires.Local().Format(time.RFC3339)
		}
		_, _ = fmt.Fprintln(stdout, line)
	}
	return 0
}

func runUseAccount(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("use", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy use ACCOUNT_NAME")
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		_, _ = fmt.Fprintln(stderr, "An account name is required.")
		return 2
	}
	store, err := authcache.NewAccountStore()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Locate authentication accounts - %v\n", err)
		return 1
	}
	account, err := store.Use(flags.Arg(0))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Select account - %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Default account is now %q.\n", account.Name)
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
	webViewer := flags.Bool("web-viewer", false, "open a live read-only capture viewer in the default browser")
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
		store, err := authcache.NewAccountStore()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Locate authentication accounts - %v\n", err)
			return 1
		}
		accounts, err := store.List()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "List authentication accounts - %v\n", err)
			return 1
		}
		account, err := store.Default()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Select default account - %v\n", err)
			return 1
		}
		firstLoginName := ""
		if len(accounts) == 0 {
			if !isInteractiveInput(os.Stdin) {
				_, _ = fmt.Fprintln(stderr, "No authentication account is configured. Run login ACCOUNT_NAME in an interactive terminal first.")
				return 1
			}
			promptReader := bufio.NewReader(os.Stdin)
			firstLoginName, err = promptAccountName(promptReader, stderr, true)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "Read account name - %v\n", err)
				return 2
			}
			_, tokenPath, pathErr := store.TokenPath(firstLoginName)
			if pathErr != nil {
				_, _ = fmt.Fprintf(stderr, "Validate account name - %v\n", pathErr)
				return 2
			}
			account = authcache.Account{Name: firstLoginName, TokenPath: tokenPath}
		}
		forceDevice := account.CacheExpired(time.Now())
		if forceDevice {
			_, _ = fmt.Fprintf(stderr, "Default account cache reached its local max-age. Starting device authentication.\n")
		}
		tokenSource, err = authcache.NewWithOptions(account.TokenPath, stderr, authcache.Options{ForceDevice: forceDevice})
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Configure authentication - %v\n", err)
			return 1
		}
		if firstLoginName != "" {
			if _, err := tokenSource.Token(); err != nil {
				_, _ = fmt.Fprintf(stderr, "Authenticate account %q - %v\n", firstLoginName, err)
				return 1
			}
			if err := store.Upsert(firstLoginName, 0, true); err != nil {
				_, _ = fmt.Fprintf(stderr, "Save account %q - %v\n", firstLoginName, err)
				return 1
			}
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
	limitations := captureLimitations(*followTransfers, *decryptResourcePacks)
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
			"web_viewer":                   strconv.FormatBool(*webViewer),
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
	var liveCaptureViewer *viewer.Viewer
	var liveViewerServer *viewer.LocalServer
	if *webViewer {
		liveCaptureViewer, err = viewer.OpenLive(recorder.Root(), viewer.Options{})
		if err != nil {
			_ = target.Close()
			_ = recorder.Close("failed", err)
			_, _ = fmt.Fprintf(stderr, "Open live capture viewer - %v\n", err)
			return 1
		}
		liveViewerServer, err = viewer.StartLocal(liveCaptureViewer)
		if err != nil {
			_ = target.Close()
			_ = recorder.Close("failed", err)
			_, _ = fmt.Fprintf(stderr, "Start live capture viewer - %v\n", err)
			return 1
		}
		defer func() {
			shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := liveViewerServer.Close(shutdownContext); err != nil {
				_, _ = fmt.Fprintf(stderr, "Stop live capture viewer - %v\n", err)
			}
		}()
		_, _ = fmt.Fprintf(stdout, "Live capture viewer %s\n", liveViewerServer.URL)
		_, _ = fmt.Fprintln(stdout, "The viewer follows complete capture events and does not run on the packet forwarding path.")
		if err := viewer.OpenBrowser(liveViewerServer.URL); err != nil {
			_, _ = fmt.Fprintf(stderr, "Open browser - %v. Open the printed URL manually.\n", err)
		}
	}

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
	_, _ = fmt.Fprintln(stdout, "Proxy forwarding stopped. Closing network connections...")
	runErr = errors.Join(runErr, target.Close())
	status := "closed"
	if runErr != nil {
		status = "failed"
	}
	_, _ = fmt.Fprintln(stdout, "Closing capture recorder...")
	closeErr := recorder.Close(status, runErr)
	if closeErr != nil {
		runErr = errors.Join(runErr, closeErr)
	}
	_, _ = fmt.Fprintln(stdout, "Capture recorder closed. Finishing automatic artifact folders...")
	artifactStatus, artifactErr := views.Finish()
	runErr = errors.Join(runErr, artifactErr)
	_, _ = fmt.Fprintf(stdout, "Automatic artifacts %s through event %d. Verifying capture...\n", artifactStatus.State, artifactStatus.LastSequence)
	verification, verifyErr := capture.Verify(recorder.Root())
	if verifyErr != nil {
		runErr = errors.Join(runErr, verifyErr)
	} else if len(verification.Issues) != 0 {
		runErr = errors.Join(runErr, fmt.Errorf("capture verification found %d issue(s)", len(verification.Issues)))
		for _, issue := range verification.Issues {
			_, _ = fmt.Fprintf(stderr, "Capture issue - %s\n", issue)
		}
	}
	if liveCaptureViewer != nil {
		if err := liveCaptureViewer.Refresh(); err != nil {
			_, _ = fmt.Fprintf(stderr, "Refresh completed capture in viewer - %v\n", err)
		}
	}
	_, _ = fmt.Fprintln(stdout, "Capture verification complete.")
	if runErr != nil {
		_, _ = fmt.Fprintf(stderr, "Proxy stopped with an error - %v\n", runErr)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Capture verified with %d events and %d blobs.\n", verification.Events, verification.Blobs)
	return 0
}

func captureLimitations(followTransfers, decryptResourcePacks bool) []string {
	limitations := []string{
		"Raw UDP datagrams and RakNet acknowledgement or retransmission frames are not captured",
		"Upstream resource-pack offer metadata is preserved; URL entries are replaced with proxy-served archives and failed retrieval remains observable",
		"Early transport payloads may not expose compression or encryption state before the connection object is available",
	}
	if followTransfers {
		limitations = append(limitations, "Transfer following is opt-in; server Transfer packets are rewritten to the local listener and each subsequent hop is captured in the same session")
	} else {
		limitations = append(limitations, "Transfer packets are recorded but automatic hop following is disabled")
	}
	if decryptResourcePacks {
		limitations = append(limitations, "Resource pack decryption supports only the documented AES-256-CFB8 contents format; unsupported variants are retained with a decrypt error")
	} else {
		limitations = append(limitations, "Encrypted resource pack archives and content keys are stored; decryption was not requested")
	}
	return limitations
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

func runView(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("view", flag.ContinueOnError)
	flags.SetOutput(stderr)
	noOpen := flags.Bool("no-open", false, "do not open the viewer in the default browser")
	live := flags.Bool("live", false, "follow a capture that may still be open")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy view [--live] [--no-open] CAPTURE_DIRECTORY")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	var captureViewer *viewer.Viewer
	var err error
	if *live {
		_, _ = fmt.Fprintln(stdout, "Indexing complete events and following the capture...")
		captureViewer, err = viewer.OpenLive(flags.Arg(0), viewer.Options{})
	} else {
		_, _ = fmt.Fprintln(stdout, "Verifying and indexing the closed capture...")
		captureViewer, err = viewer.Open(flags.Arg(0), viewer.Options{})
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Open capture viewer - %v\n", err)
		return 1
	}
	server, err := viewer.StartLocal(captureViewer)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Start capture viewer - %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Capture viewer %s\n", server.URL)
	_, _ = fmt.Fprintln(stdout, "The viewer is read-only and bound to this computer. Press Ctrl+C to stop it.")
	if !*noOpen {
		if err := viewer.OpenBrowser(server.URL); err != nil {
			_, _ = fmt.Fprintf(stderr, "Open browser - %v. Open the printed URL manually.\n", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-server.Done():
		err := server.Err()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Capture viewer stopped with an error - %v\n", err)
			return 1
		}
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := server.Close(shutdownContext)
		cancel()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Stop capture viewer - %v\n", err)
			return 1
		}
	}
	_, _ = fmt.Fprintln(stdout, "Capture viewer stopped.")
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
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy login [--max-age 7d] [ACCOUNT_NAME]")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy accounts")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy use ACCOUNT_NAME")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy logout [ACCOUNT_NAME]")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy verify CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy inspect [filters] CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy view [--live] [--no-open] CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy analyze CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy explain CAPTURE_DIRECTORY")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy export CAPTURE_DIRECTORY OUTPUT.bdpcap")
	_, _ = fmt.Fprintln(writer, "  bedrock-debug-proxy version")
}
