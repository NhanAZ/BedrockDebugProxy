package packetview

import (
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
)

const maxSafeJSONInteger = int64(1<<53 - 1)

type Options struct {
	MaxDepth           int
	MaxCollectionItems int
	BinaryPreviewBytes int
}

type Notice struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

type encoder struct {
	options Options
	notices []Notice
	active  map[visit]struct{}
}

type visit struct {
	typeName reflect.Type
	pointer  uintptr
}

func Encode(value any, options Options) (data json.RawMessage, notices []Notice, err error) {
	if options.MaxDepth <= 0 {
		options.MaxDepth = 128
	}
	if options.BinaryPreviewBytes < 0 {
		options.BinaryPreviewBytes = 0
	}
	e := &encoder{options: options, active: map[visit]struct{}{}}
	defer func() {
		if recovered := recover(); recovered != nil {
			data = nil
			notices = e.notices
			err = fmt.Errorf("normalize decoded value: %v", recovered)
		}
	}()
	normalized := e.value(reflect.ValueOf(value), "$", 0)
	data, err = json.Marshal(normalized)
	if err != nil {
		return nil, e.notices, fmt.Errorf("encode normalized value: %w", err)
	}
	return data, e.notices, nil
}

func (e *encoder) value(value reflect.Value, path string, depth int) any {
	if !value.IsValid() {
		return nil
	}
	if depth >= e.options.MaxDepth {
		e.notice(path, "max_depth", fmt.Sprintf("value exceeded maximum depth %d", e.options.MaxDepth))
		return map[string]any{"$truncated": "max_depth", "$type": value.Type().String()}
	}
	if value.CanInterface() {
		if marshaler, ok := value.Interface().(encoding.TextMarshaler); ok {
			text, err := marshaler.MarshalText()
			if err == nil {
				return string(text)
			}
			e.notice(path, "text_marshal_error", err.Error())
		}
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return e.value(value.Elem(), path, depth+1)
	case reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		key := visit{typeName: value.Type(), pointer: value.Pointer()}
		if _, exists := e.active[key]; exists {
			e.notice(path, "cycle", "cyclic pointer was replaced with a marker")
			return map[string]any{"$cycle": value.Type().String()}
		}
		e.active[key] = struct{}{}
		defer delete(e.active, key)
		return e.value(value.Elem(), path, depth+1)
	case reflect.Bool:
		return value.Bool()
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := value.Int()
		if integer > maxSafeJSONInteger || integer < -maxSafeJSONInteger {
			e.notice(path, "large_integer", "integer was encoded as a decimal string to preserve precision")
			return map[string]any{"$integer": strconv.FormatInt(integer, 10), "$type": value.Type().String()}
		}
		return integer
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		integer := value.Uint()
		if integer > uint64(maxSafeJSONInteger) {
			e.notice(path, "large_integer", "integer was encoded as a decimal string to preserve precision")
			return map[string]any{"$integer": strconv.FormatUint(integer, 10), "$type": value.Type().String()}
		}
		return integer
	case reflect.Float32, reflect.Float64:
		f := value.Float()
		if math.IsNaN(f) {
			e.notice(path, "non_finite_float", "NaN was encoded as a string")
			return "NaN"
		}
		if math.IsInf(f, 1) {
			e.notice(path, "non_finite_float", "positive infinity was encoded as a string")
			return "+Infinity"
		}
		if math.IsInf(f, -1) {
			e.notice(path, "non_finite_float", "negative infinity was encoded as a string")
			return "-Infinity"
		}
		return f
	case reflect.Complex64, reflect.Complex128:
		complexValue := value.Complex()
		e.notice(path, "complex_number", "complex number was encoded as real and imaginary parts")
		return map[string]any{"real": real(complexValue), "imaginary": imag(complexValue)}
	case reflect.Struct:
		return e.structure(value, path, depth)
	case reflect.Slice:
		if value.IsNil() {
			return nil
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return e.binary(value.Bytes())
		}
		return e.collection(value, path, depth)
	case reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			bytesValue := make([]byte, value.Len())
			for index := range bytesValue {
				bytesValue[index] = byte(value.Index(index).Uint())
			}
			return e.binary(bytesValue)
		}
		return e.collection(value, path, depth)
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		return e.mapValue(value, path, depth)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		e.notice(path, "unsupported_type", fmt.Sprintf("%s value was replaced with a marker", value.Kind()))
		return map[string]any{"$unsupported": value.Type().String()}
	default:
		e.notice(path, "unsupported_type", fmt.Sprintf("%s value was replaced with a marker", value.Kind()))
		return map[string]any{"$unsupported": value.Type().String()}
	}
}

func (e *encoder) structure(value reflect.Value, path string, depth int) any {
	typeInfo := value.Type()
	result := make(map[string]any, value.NumField()+1)
	result["$type"] = typeInfo.String()
	for index := 0; index < value.NumField(); index++ {
		fieldInfo := typeInfo.Field(index)
		if !fieldInfo.IsExported() {
			continue
		}
		name := fieldInfo.Name
		if tag := fieldInfo.Tag.Get("json"); tag != "" {
			tagName := tag
			if comma := indexByte(tagName, ','); comma >= 0 {
				tagName = tagName[:comma]
			}
			if tagName == "-" {
				continue
			}
			if tagName != "" {
				name = tagName
			}
		}
		result[name] = e.value(value.Field(index), fieldPath(path, name), depth+1)
	}
	return result
}

func (e *encoder) collection(value reflect.Value, path string, depth int) any {
	length := value.Len()
	limit := length
	if e.options.MaxCollectionItems > 0 && limit > e.options.MaxCollectionItems {
		limit = e.options.MaxCollectionItems
		e.notice(path, "collection_truncated", fmt.Sprintf("kept %d of %d items", limit, length))
	}
	items := make([]any, 0, limit)
	for index := 0; index < limit; index++ {
		items = append(items, e.value(value.Index(index), indexPath(path, index), depth+1))
	}
	if limit == length {
		return items
	}
	return map[string]any{"$items": items, "$total": length, "$truncated": length - limit}
}

func (e *encoder) mapValue(value reflect.Value, path string, depth int) any {
	keys := value.MapKeys()
	limit := len(keys)
	if e.options.MaxCollectionItems > 0 && limit > e.options.MaxCollectionItems {
		limit = e.options.MaxCollectionItems
		e.notice(path, "collection_truncated", fmt.Sprintf("kept %d of %d map entries", limit, len(keys)))
	}
	if value.Type().Key().Kind() == reflect.String {
		names := make([]string, 0, len(keys))
		for _, key := range keys {
			names = append(names, key.String())
		}
		sort.Strings(names)
		if limit < len(names) {
			names = names[:limit]
		}
		result := make(map[string]any, len(names)+1)
		for _, name := range names {
			result[name] = e.value(value.MapIndex(reflect.ValueOf(name).Convert(value.Type().Key())), fieldPath(path, name), depth+1)
		}
		if limit < len(keys) {
			result["$truncated"] = len(keys) - limit
		}
		return result
	}
	type entry struct {
		KeyText string
		Key     reflect.Value
	}
	entries := make([]entry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, entry{KeyText: key.Type().String() + ":" + fmt.Sprint(key.Interface()), Key: key})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].KeyText < entries[j].KeyText })
	if limit < len(entries) {
		entries = entries[:limit]
	}
	result := make([]any, 0, len(entries))
	for index, item := range entries {
		result = append(result, map[string]any{
			"key":   e.value(item.Key, indexPath(path, index)+".key", depth+1),
			"value": e.value(value.MapIndex(item.Key), indexPath(path, index)+".value", depth+1),
		})
	}
	wrapped := map[string]any{"$map": result}
	if limit < len(keys) {
		wrapped["$truncated"] = len(keys) - limit
	}
	return wrapped
}

func (e *encoder) binary(value []byte) any {
	digest := sha256.Sum256(value)
	previewLength := e.options.BinaryPreviewBytes
	if previewLength > len(value) {
		previewLength = len(value)
	}
	result := map[string]any{
		"size":   len(value),
		"sha256": hex.EncodeToString(digest[:]),
	}
	if previewLength != 0 {
		result["preview_hex"] = hex.EncodeToString(value[:previewLength])
		result["preview_bytes"] = previewLength
	}
	if previewLength < len(value) {
		result["omitted_bytes"] = len(value) - previewLength
	}
	return map[string]any{"$binary": result}
}

func (e *encoder) notice(path, kind, message string) {
	e.notices = append(e.notices, Notice{Path: path, Kind: kind, Message: message})
}

func fieldPath(parent, field string) string {
	return parent + "." + strconv.Quote(field)
}

func indexPath(parent string, index int) string {
	return parent + "[" + strconv.Itoa(index) + "]"
}

func indexByte(value string, target byte) int {
	for index := 0; index < len(value); index++ {
		if value[index] == target {
			return index
		}
	}
	return -1
}
