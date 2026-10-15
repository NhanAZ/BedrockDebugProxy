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
	"syscall"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/golem"
)

func runGolem(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("golem", flag.ContinueOnError)
	flags.SetOutput(stderr)
	directory := flags.String("data", "", "GolemAgent plugin data directory on the development server")
	input := flags.String("input", "-", "JSON Lines scenario file, or - for stdin")
	request := flags.String("request", "", "one JSON request instead of --input")
	timeout := flags.Duration("timeout", 30*time.Second, "wall-clock deadline per request, maximum 10m")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage - bedrock-debug-proxy golem --data DIRECTORY [--request JSON | --input FILE] [--timeout 30s]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	inputSet := false
	requestSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "input" {
			inputSet = true
		}
		if f.Name == "request" {
			requestSet = true
		}
	})
	if flags.NArg() != 0 || *directory == "" || *timeout <= 0 || *timeout > 10*time.Minute || (requestSet && (inputSet || *request == "")) {
		flags.Usage()
		return 2
	}
	var reader io.Reader = os.Stdin
	if *request != "" {
		reader = bytes.NewBufferString(*request)
	} else if *input != "-" {
		file, err := os.Open(*input)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		defer func() { _ = file.Close() }()
		reader = file
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if closer, ok := reader.(io.Closer); ok {
		stopRead := context.AfterFunc(ctx, func() { _ = closer.Close() })
		defer stopRead()
	}
	return runGolemInput(ctx, *directory, reader, stdout, stderr, *timeout)
}

func runGolemInput(ctx context.Context, directory string, input io.Reader, stdout, stderr io.Writer, timeout time.Duration) int {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), golem.MaxMessageBytes)
	encoder := json.NewEncoder(stdout)
	count := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var request golem.Request
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			_, _ = fmt.Fprintf(stderr, "Invalid Golem request - %v\n", err)
			return 2
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) || request.Op == "" {
			_, _ = fmt.Fprintln(stderr, "Each line must contain exactly one request with an op.")
			return 2
		}
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		response, err := golem.Call(callCtx, directory, request)
		cancel()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Golem request %q - %v\n", request.ID, err)
			return 1
		}
		if err := encoder.Encode(response); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		count++
		if !response.OK {
			return 1
		}
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Read Golem scenario (1 MiB per line limit) - %v\n", err)
		return 2
	}
	if count == 0 {
		_, _ = fmt.Fprintln(stderr, "No Golem requests supplied.")
		return 2
	}
	return 0
}
