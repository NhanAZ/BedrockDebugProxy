package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/authcache"
	"github.com/NhanAZ/BedrockDebugProxy/internal/experience"
	"github.com/NhanAZ/BedrockDebugProxy/internal/golemnet"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"golang.org/x/oauth2"
)

const maxNetworkScenarioLine = 1 << 20

type networkScenarioRequest struct {
	Op       string          `json:"op"`
	ID       string          `json:"id,omitempty"`
	Message  string          `json:"message,omitempty"`
	Index    *int            `json:"index,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Duration string          `json:"duration,omitempty"`
}

type networkScenarioResponse struct {
	Schema string `json:"schema"`
	ID     string `json:"id,omitempty"`
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func runGolemNetwork(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("golem network", flag.ContinueOnError)
	flags.SetOutput(stderr)
	server := ""
	flags.StringVar(&server, "server", "", "destination HOST:PORT or experience:NAME")
	flags.StringVar(&server, "upstream", "", "alias for --server")
	authMode := flags.String("auth", "device", "upstream authentication mode - device (cached) or none")
	input := flags.String("input", "-", "JSON Lines scenario file, or - for stdin")
	report := flags.String("report", "", "write the private local session report to this new JSON file")
	requestTimeout := flags.Duration("timeout", 30*time.Second, "deadline per scenario request, maximum 10m")
	connectTimeout := flags.Duration("connect-timeout", 15*time.Minute, "deadline for login, packs, and spawn")
	followTransfers := flags.Bool("follow-transfers", true, "follow server Transfer packets and spawn at the next hop")
	maxHops := flags.Int("max-hops", 64, "maximum transfer hops in one session")
	name := flags.String("name", "BDP-Golem", "offline identity name for unauthenticated direct servers")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy golem network --server HOST:PORT|experience:NAME [options]")
		flags.PrintDefaults()
		_, _ = fmt.Fprintln(stderr, "Scenario operations - wait_form, current_form, select_form, submit_form, chat, wait, observe, report, quit")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || strings.TrimSpace(server) == "" || *requestTimeout <= 0 || *requestTimeout > 10*time.Minute || *connectTimeout <= 0 || *connectTimeout > 30*time.Minute || *maxHops < 1 || *maxHops > 64 {
		flags.Usage()
		return 2
	}
	if *authMode != "device" && *authMode != "none" {
		_, _ = fmt.Fprintln(stderr, "Unsupported --auth. Use device or none.")
		return 2
	}

	tokenSource, accountName, err := configureGolemNetworkAuth(*authMode, stderr)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Configure authentication - %v\n", err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	target, err := experience.Resolve(ctx, server, tokenSource, nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Resolve server - %v\n", err)
		return 1
	}
	defer func() { _ = target.Close() }()
	if accountName != "" {
		_, _ = fmt.Fprintf(stderr, "Using Microsoft account %q.\n", accountName)
	}
	identity := login.IdentityData{}
	if *authMode == "none" {
		identity = login.IdentityData{Identity: uuid.NewString(), DisplayName: *name}
	}
	client, err := golemnet.New(golemnet.Config{
		Address:         target.Address,
		Transport:       target.Transport,
		Network:         target.Network,
		TokenSource:     target.TokenSource,
		Identity:        identity,
		ConnectTimeout:  *connectTimeout,
		FollowTransfers: *followTransfers,
		MaxHops:         *maxHops,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Configure network client - %v\n", err)
		return 2
	}
	err = client.Start(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Start network Golem - %v\n", err)
		_ = client.Close()
		return 1
	}
	defer func() { _ = client.Close() }()
	if err := writeNetworkResponse(stdout, networkScenarioResponse{
		Schema: golemnet.Schema,
		ID:     "connect",
		OK:     true,
		Result: map[string]any{"transport": target.Transport, "name": target.Name, "authenticated": client.Snapshot().Authenticated},
	}); err != nil {
		_, _ = fmt.Fprintf(stderr, "Write connection response - %v\n", err)
		return 1
	}

	reader, closeReader, err := openScenarioInput(*input)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Open scenario input - %v\n", err)
		return 2
	}
	if closeReader != nil {
		defer func() { _ = closeReader.Close() }()
	}
	status := runNetworkScenario(ctx, client, reader, stdout, stderr, *requestTimeout)
	closeErr := client.Close()
	if status == 0 && closeErr != nil {
		_, _ = fmt.Fprintf(stderr, "Close network Golem - %v\n", closeErr)
		status = 1
	}
	if *report != "" {
		if err := writeNetworkReport(*report, client.Snapshot()); err != nil {
			_, _ = fmt.Fprintf(stderr, "Write network Golem report - %v\n", err)
			status = 1
		}
	}
	return status
}

func configureGolemNetworkAuth(mode string, stderr io.Writer) (oauth2.TokenSource, string, error) {
	if mode == "none" {
		return nil, "", nil
	}
	store, err := authcache.NewAccountStore()
	if err != nil {
		return nil, "", err
	}
	accounts, err := store.List()
	if err != nil {
		return nil, "", err
	}
	account, err := store.Default()
	firstLoginName := ""
	if err != nil {
		if len(accounts) != 0 {
			return nil, "", err
		}
		if !isInteractiveInput(os.Stdin) {
			return nil, "", errors.New("no authentication account is configured. Run login ACCOUNT_NAME in an interactive terminal first")
		}
		reader := bufio.NewReader(os.Stdin)
		firstLoginName, err = promptAccountName(reader, stderr, true)
		if err != nil {
			return nil, "", err
		}
		_, tokenPath, pathErr := store.TokenPath(firstLoginName)
		if pathErr != nil {
			return nil, "", pathErr
		}
		account = authcache.Account{Name: firstLoginName, TokenPath: tokenPath}
	}
	if account.TokenPath == "" {
		return nil, "", errors.New("selected authentication account has no token path")
	}
	source, err := authcache.New(account.TokenPath, stderr)
	if err != nil {
		return nil, "", err
	}
	if firstLoginName != "" {
		if _, err := source.Token(); err != nil {
			return nil, "", err
		}
		if err := store.Upsert(firstLoginName, 0, true); err != nil {
			return nil, "", err
		}
	}
	return source, account.Name, nil
}

func openScenarioInput(path string) (io.Reader, io.Closer, error) {
	if path == "-" {
		return os.Stdin, nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return file, file, nil
}

func runNetworkScenario(ctx context.Context, client *golemnet.Client, input io.Reader, stdout, stderr io.Writer, timeout time.Duration) int {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), maxNetworkScenarioLine)
	decoderOutput := json.NewEncoder(stdout)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var request networkScenarioRequest
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			_, _ = fmt.Fprintf(stderr, "Invalid network Golem request - %v\n", err)
			return 2
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			_, _ = fmt.Fprintln(stderr, "Invalid network Golem request - each line must contain exactly one JSON object")
			return 2
		}
		if err := validateNetworkScenarioRequest(request); err != nil {
			_, _ = fmt.Fprintf(stderr, "Invalid network Golem request - %v\n", err)
			return 2
		}
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		result, quit, err := executeNetworkScenario(callCtx, client, request)
		cancel()
		response := networkScenarioResponse{Schema: golemnet.Schema, ID: request.ID, OK: err == nil, Result: result}
		if err != nil {
			response.Error = err.Error()
		}
		if encodeErr := decoderOutput.Encode(response); encodeErr != nil {
			_, _ = fmt.Fprintf(stderr, "Write network Golem response - %v\n", encodeErr)
			return 1
		}
		if err != nil {
			return 1
		}
		if quit {
			return 0
		}
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Read network Golem scenario - %v\n", err)
		return 1
	}
	return 0
}

func validateNetworkScenarioRequest(request networkScenarioRequest) error {
	if strings.TrimSpace(request.Op) == "" {
		return errors.New("op is required")
	}
	if len(request.ID) > 256 || len(request.Message) > 64<<10 || len(request.Duration) > 64 || len(request.Response) > 1<<20 {
		return errors.New("request field exceeds size limit")
	}
	switch request.Op {
	case "wait_form", "current_form", "select_form", "submit_form", "chat", "wait", "observe", "report", "quit":
		return nil
	default:
		return fmt.Errorf("unsupported op %q", request.Op)
	}
}

func executeNetworkScenario(ctx context.Context, client *golemnet.Client, request networkScenarioRequest) (any, bool, error) {
	switch request.Op {
	case "wait_form":
		form, err := client.WaitForForm(ctx)
		return form, false, err
	case "current_form":
		form, ok := client.CurrentForm()
		if !ok {
			return nil, false, golemnet.ErrNoCurrentForm
		}
		return form, false, nil
	case "select_form":
		if request.Index == nil {
			return nil, false, errors.New("select_form requires index")
		}
		return map[string]any{"selected": *request.Index}, false, client.SelectForm(ctx, *request.Index)
	case "submit_form":
		return map[string]any{"submitted": true}, false, client.SubmitForm(ctx, request.Response)
	case "chat":
		return map[string]any{"sent": true}, false, client.Chat(request.Message)
	case "wait":
		duration, err := parseScenarioDuration(request.Duration)
		if err != nil {
			return nil, false, err
		}
		return map[string]any{"waited": duration.String()}, false, client.Wait(ctx, duration)
	case "observe":
		duration, err := parseScenarioDuration(request.Duration)
		if err != nil {
			return nil, false, err
		}
		if err := client.Wait(ctx, duration); err != nil {
			return nil, false, err
		}
		return client.Snapshot(), false, nil
	case "report":
		return client.Snapshot(), false, nil
	case "quit":
		return map[string]any{"closed": true}, true, client.Close()
	default:
		return nil, false, fmt.Errorf("unsupported op %q", request.Op)
	}
}

func parseScenarioDuration(value string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 1 * time.Second, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 || duration > 10*time.Minute {
		return 0, errors.New("duration must be positive, valid, and at most 10m")
	}
	return duration, nil
}

func writeNetworkResponse(writer io.Writer, response networkScenarioResponse) error {
	return json.NewEncoder(writer).Encode(response)
}

func writeNetworkReport(path string, report golemnet.Report) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("report path is empty")
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create report directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create report file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if err := json.NewEncoder(file).Encode(report); err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync report: %w", err)
	}
	return file.Close()
}
