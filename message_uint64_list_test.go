package hyperpb_test

import (
	"slices"
	"testing"

	"buf.build/go/hyperpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Keep the result at an interface boundary, as the production balance cache does.
var uint64ListSink hyperpb.Uint64List

func TestUint64ListBorrowedZeroAllocAndReuse(t *testing.T) {
	for _, packed := range []bool{false, true} {
		label := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		u64, fixed := descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64
		file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
			Name: proto.String("uint64_list.proto"), Syntax: proto.String("proto3"),
			MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Balances"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("varints"), Number: proto.Int32(1), Label: &label, Type: &u64, Options: &descriptorpb.FieldOptions{Packed: proto.Bool(packed)}},
				{Name: proto.String("fixed"), Number: proto.Int32(2), Label: &label, Type: &fixed, Options: &descriptorpb.FieldOptions{Packed: proto.Bool(packed)}},
			}}},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		desc := file.Messages().Get(0)
		typ := hyperpb.CompileMessageDescriptor(desc)
		values := [][]uint64{nil, {1, 2, 3}, {0, 127, 128, 1 << 32, ^uint64(0)}, {9}}
		wires := make([][]byte, len(values))
		for i, vals := range values {
			m := dynamicpb.NewMessage(desc)
			for f := 0; f < 2; f++ {
				list := m.Mutable(desc.Fields().Get(f)).List()
				for _, value := range vals {
					list.Append(protoreflect.ValueOfUint64(value))
				}
			}
			wires[i], err = proto.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, alias := range []bool{false, true} {
			var shared hyperpb.Shared
			for round := 0; round < 8; round++ {
				for i, wire := range wires {
					msg := shared.NewMessage(typ)
					if err := msg.Unmarshal(wire, hyperpb.WithAllowAlias(alias)); err != nil {
						t.Fatal(err)
					}
					for f := 0; f < 2; f++ {
						list := hyperpb.Uint64ListByIndex(msg, f)
						if list == nil {
							t.Fatal("missing typed view")
						}
						if list.Len() != len(values[i]) {
							t.Fatalf("round=%d field=%d length=%d", round, f, list.Len())
						}
						for j, value := range values[i] {
							if list.Get(j) != value {
								t.Fatalf("round=%d field=%d index=%d stale value", round, f, j)
							}
						}
						var copied [8]uint64
						if !slices.Equal(list.Copy(copied[:0]), values[i]) {
							t.Fatal("copy mismatch")
						}
						if got := testing.AllocsPerRun(100, func() {
							uint64ListSink = hyperpb.Uint64ListByIndex(msg, f)
						}); got != 0 {
							t.Fatalf("packed=%v alias=%v field=%d allocs=%v", packed, alias, f, got)
						}
					}
					if hyperpb.Uint64ListByIndex(msg, -1) != nil || hyperpb.Uint64ListByIndex(msg, 2) != nil {
						t.Fatal("invalid index")
					}
					uint64ListSink = nil
					// Old borrowed lists are intentionally not used after Free.
					shared.Free()
				}
			}
		}
	}
}
