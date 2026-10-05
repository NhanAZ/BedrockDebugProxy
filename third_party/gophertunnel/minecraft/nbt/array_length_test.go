package nbt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestFixedWidthArraysRejectOverflowAndTruncation(t *testing.T) {
	for _, encoding := range []struct {
		name  string
		value Encoding
		order binary.ByteOrder
	}{{"little endian", LittleEndian, binary.LittleEndian}, {"big endian", BigEndian, binary.BigEndian}} {
		for _, test := range []struct {
			name  string
			count uint32
			wide  bool
		}{{"negative int32", 0xffffffff, false}, {"overflow int32", 0x40000000, false}, {"overflow int64", 0x20000000, true}, {"truncated int32", 2, false}, {"truncated int64", 2, true}} {
			t.Run(encoding.name+"/"+test.name, func(t *testing.T) {
				data := make([]byte, 4)
				encoding.order.PutUint32(data, test.count)
				for _, stream := range []bool{false, true} {
					var source io.Reader = bytes.NewBuffer(data)
					if stream {
						source = struct{ io.Reader }{source}
					}
					reader := newOffsetReader(source)
					var err error
					if test.wide {
						_, err = encoding.value.Int64Slice(reader)
					} else {
						_, err = encoding.value.Int32Slice(reader)
					}
					var overrun BufferOverrunError
					if !errors.As(err, &overrun) {
						t.Fatalf("stream=%t: error = %v, want BufferOverrunError", stream, err)
					}
				}
			})
		}
	}
}

func TestFixedWidthArraysPreserveValuesAndOffsets(t *testing.T) {
	for _, encoding := range []struct {
		value Encoding
		order binary.ByteOrder
	}{{LittleEndian, binary.LittleEndian}, {BigEndian, binary.BigEndian}} {
		for _, stream := range []bool{false, true} {
			for _, empty := range []bool{false, true} {
				values32 := []int32{-123456, 0, 123456}
				values64 := []int64{-1 << 40, 0, 1 << 40}
				if empty {
					values32, values64 = []int32{}, []int64{}
				}
				var data bytes.Buffer
				for _, value := range []any{int32(len(values32)), values32, int32(len(values64)), values64} {
					if err := binary.Write(&data, encoding.order, value); err != nil {
						t.Fatal(err)
					}
				}
				length := data.Len()
				data.WriteByte(0x7f)
				var source io.Reader = &data
				if stream {
					source = struct{ io.Reader }{source}
				}
				reader := newOffsetReader(source)
				got32, err := encoding.value.Int32Slice(reader)
				if err != nil {
					t.Fatal(err)
				}
				got64, err := encoding.value.Int64Slice(reader)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got32, values32) || !reflect.DeepEqual(got64, values64) || reader.off != int64(length) {
					t.Fatalf("values or offsets changed: %v, %v, %d", got32, got64, reader.off)
				}
				if next, err := reader.ReadByte(); err != nil || next != 0x7f {
					t.Fatalf("array consumed trailing data: %x, %v", next, err)
				}
			}
		}
	}
}
