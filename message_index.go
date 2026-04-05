package hyperpb

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"buf.build/go/hyperpb/internal/xprotoreflect"
)

type indexGetter interface {
	GetByIndexUnchecked(int) protoreflect.Value
}

type indexedMessageGetter interface {
	GetMessageByIndexUnchecked(int) protoreflect.Message
}

type indexedListGetter interface {
	GetListByIndexUnchecked(int) protoreflect.List
}

type indexedStringGetter interface {
	GetStringByIndexUnchecked(int) string
}

type indexedUint64Getter interface {
	GetUint64ByIndexUnchecked(int) uint64
}

type indexedUint64ListGetter interface {
	GetUint64ListByIndexUnchecked(int) Uint64List
}

type indexedInt64Getter interface {
	GetInt64ByIndexUnchecked(int) int64
}

type messageListGetter interface {
	GetMessage(int) protoreflect.Message
}

type reflectUint64List struct {
	raw protoreflect.List
}

func (r reflectUint64List) Len() int {
	if r.raw == nil {
		return 0
	}
	return r.raw.Len()
}

func (r reflectUint64List) Get(n int) uint64 {
	if r.raw == nil || n < 0 || n >= r.raw.Len() {
		return 0
	}
	return r.raw.Get(n).Uint()
}

func (r reflectUint64List) Copy(out []uint64) []uint64 {
	if r.raw == nil || r.raw.Len() == 0 {
		return nil
	}
	for i := 0; i < r.raw.Len(); i++ {
		out = append(out, r.raw.Get(i).Uint())
	}
	return out
}

// GetByIndex returns a field by raw descriptor index, using hyperpb's fast
// path when the message supports it.
func GetByIndex(msg protoreflect.Message, index int) protoreflect.Value {
	if msg == nil || index < 0 {
		return protoreflect.Value{}
	}
	if fast, ok := msg.(indexGetter); ok {
		return fast.GetByIndexUnchecked(index)
	}
	if !msg.IsValid() {
		return protoreflect.Value{}
	}
	fields := msg.Descriptor().Fields()
	if index >= fields.Len() {
		return protoreflect.Value{}
	}
	return msg.Get(fields.Get(index))
}

// ListMessageAt returns a nested message from a protoreflect.List.
func ListMessageAt(list protoreflect.List, index int) protoreflect.Message {
	if list == nil || index < 0 || index >= list.Len() {
		return nil
	}
	if fast, ok := list.(messageListGetter); ok {
		msg := fast.GetMessage(index)
		if !msg.IsValid() {
			return nil
		}
		return msg
	}
	v := list.Get(index)
	if !v.IsValid() {
		return nil
	}
	msg := v.Message()
	if !msg.IsValid() {
		return nil
	}
	return msg
}

// MessageByIndex returns a nested message field by raw descriptor index, or
// nil if the field is unset.
func MessageByIndex(msg protoreflect.Message, index int) protoreflect.Message {
	if msg == nil || index < 0 {
		return nil
	}
	if fast, ok := msg.(indexedMessageGetter); ok {
		nested := fast.GetMessageByIndexUnchecked(index)
		if !nested.IsValid() {
			return nil
		}
		return nested
	}
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return nil
	}
	nested := xprotoreflect.GetMessage[protoreflect.Message](v)
	if !nested.IsValid() {
		return nil
	}
	return nested
}

// ListByIndex returns a repeated field by raw descriptor index.
func ListByIndex(msg protoreflect.Message, index int) protoreflect.List {
	if msg == nil || index < 0 {
		return nil
	}
	if fast, ok := msg.(indexedListGetter); ok {
		return fast.GetListByIndexUnchecked(index)
	}
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return nil
	}
	return xprotoreflect.List(v)
}

// BytesByIndex returns a bytes field by raw descriptor index.
func BytesByIndex(msg protoreflect.Message, index int) []byte {
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return nil
	}
	return v.Bytes()
}

// StringByIndex returns a string field by raw descriptor index.
func StringByIndex(msg protoreflect.Message, index int) string {
	if msg == nil || index < 0 {
		return ""
	}
	if fast, ok := msg.(indexedStringGetter); ok {
		return fast.GetStringByIndexUnchecked(index)
	}
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return ""
	}
	return xprotoreflect.GetString(v)
}

// Uint64ByIndex returns a uint-like field by raw descriptor index.
func Uint64ByIndex(msg protoreflect.Message, index int) uint64 {
	if msg == nil || index < 0 {
		return 0
	}
	if fast, ok := msg.(indexedUint64Getter); ok {
		return fast.GetUint64ByIndexUnchecked(index)
	}
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return 0
	}
	return xprotoreflect.GetRawInt(v)
}

// Uint64ListByIndex returns a repeated uint64/fixed64 field by raw
// descriptor index, using hyperpb's typed fast path when available.
func Uint64ListByIndex(msg protoreflect.Message, index int) Uint64List {
	if msg == nil || index < 0 {
		return nil
	}
	if fast, ok := msg.(indexedUint64ListGetter); ok {
		return fast.GetUint64ListByIndexUnchecked(index)
	}
	list := ListByIndex(msg, index)
	if list == nil {
		return nil
	}
	return reflectUint64List{raw: list}
}

// Int64ByIndex returns an int-like field by raw descriptor index.
func Int64ByIndex(msg protoreflect.Message, index int) int64 {
	if msg == nil || index < 0 {
		return 0
	}
	if fast, ok := msg.(indexedInt64Getter); ok {
		return fast.GetInt64ByIndexUnchecked(index)
	}
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return 0
	}
	return int64(xprotoreflect.GetRawInt(v))
}

// BoolByIndex returns a bool field by raw descriptor index.
func BoolByIndex(msg protoreflect.Message, index int) bool {
	v := GetByIndex(msg, index)
	if !v.IsValid() {
		return false
	}
	return v.Bool()
}
