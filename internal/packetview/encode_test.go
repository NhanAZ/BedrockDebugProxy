package packetview

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

type sample struct {
	Name   string
	Bytes  []byte
	Values map[int]string
	Next   *sample
	hidden string
}

func TestEncodeProducesTypedJSONSafeView(t *testing.T) {
	value := &sample{
		Name:   "packet",
		Bytes:  []byte{0xde, 0xad, 0xbe, 0xef},
		Values: map[int]string{2: "two", 1: "one"},
		hidden: "must not be captured",
	}
	value.Next = value
	encoded, notices, err := Encode(value, Options{BinaryPreviewBytes: 2})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !json.Valid(encoded) {
		t.Fatalf("invalid JSON: %s", encoded)
	}
	text := string(encoded)
	for _, expected := range []string{`"$type":"packetview.sample"`, `"preview_hex":"dead"`, `"omitted_bytes":2`, `"$map"`, `"$cycle"`} {
		if !strings.Contains(text, expected) {
			t.Errorf("encoded JSON does not contain %s: %s", expected, text)
		}
	}
	if strings.Contains(text, value.hidden) {
		t.Fatalf("encoded JSON exposes an unexported field: %s", text)
	}
	if len(notices) != 1 || notices[0].Kind != "cycle" {
		t.Fatalf("notices = %#v", notices)
	}
}

func TestEncodeMarksTruncationAndNonFiniteFloats(t *testing.T) {
	value := struct {
		Items []int
		Value float64
		Large uint64
	}{Items: []int{1, 2, 3}, Value: math.NaN(), Large: 1 << 63}
	encoded, notices, err := Encode(value, Options{MaxCollectionItems: 2})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `"$truncated":1`) || !strings.Contains(text, `"Value":"NaN"`) || !strings.Contains(text, `"$integer":"9223372036854775808"`) {
		t.Fatalf("encoded JSON = %s", text)
	}
	if len(notices) != 3 {
		t.Fatalf("notices = %#v", notices)
	}
}
