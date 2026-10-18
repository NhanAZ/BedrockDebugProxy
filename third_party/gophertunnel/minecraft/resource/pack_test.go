package resource

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadPreservesArchiveReadError(t *testing.T) {
	archive := streamTestArchive(t)
	failure := errors.New("synthetic archive stream failure")
	for _, test := range []struct {
		name string
		data []byte
	}{
		{name: "before archive"},
		{name: "truncated archive", data: archive[:len(archive)/2]},
		{name: "after complete archive", data: archive},
	} {
		t.Run(test.name, func(t *testing.T) {
			pack, err := Read(io.MultiReader(bytes.NewReader(test.data), failingPackReader{failure}))
			if !errors.Is(err, failure) {
				t.Fatalf("Read error = %v, want original stream failure", err)
			}
			if pack != nil {
				t.Fatal("accepted an archive from a failed stream")
			}
		})
	}
}

func TestReadPreservesSuccessfulArchiveBytes(t *testing.T) {
	archive := streamTestArchive(t)
	pack, err := Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	got := make([]byte, pack.Len())
	if _, err := pack.ReadAt(got, 0); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, archive) {
		t.Fatal("archive bytes changed after reading")
	}
}

type failingPackReader struct{ err error }

func (r failingPackReader) Read([]byte) (int, error) { return 0, r.err }

func streamTestArchive(t *testing.T) []byte {
	t.Helper()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	entry, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(entry, `{"format_version":2,"header":{"name":"Synthetic stream fixture","description":"Synthetic test","uuid":"11111111-1111-1111-1111-111111111111","version":[1,0,0],"min_engine_version":[1,20,0]},"modules":[{"type":"resources","uuid":"22222222-2222-2222-2222-222222222222","version":[1,0,0]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
