package hyperpb_test

import (
	"testing"

	"buf.build/go/hyperpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestHasByIndexV1ConfigPresence(t *testing.T) {
	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	u64 := descriptorpb.FieldDescriptorProto_TYPE_UINT64
	message := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name: proto.String("v1_presence.proto"), Syntax: proto.String("proto3"), Package: proto.String("presence"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Config"), OneofDecl: []*descriptorpb.OneofDescriptorProto{{Name: proto.String("_priority_fee")}},
			Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("priority_fee"), Number: proto.Int32(1), Label: &optional, Type: &u64, Proto3Optional: proto.Bool(true), OneofIndex: proto.Int32(0)}},
		}, {
			Name:  proto.String("Message"),
			Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("config"), Number: proto.Int32(7), Label: &optional, Type: &message, TypeName: proto.String(".presence.Config")}},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	desc := file.Messages().ByName("Message")
	typeOf := hyperpb.CompileMessageDescriptor(desc)
	for _, tc := range []struct {
		name        string
		wire        []byte
		config, fee bool
		value       uint64
	}{
		{"absent", nil, false, false, 0},
		{"empty", []byte{0x3a, 0}, true, false, 0},
		{"zero", []byte{0x3a, 2, 8, 0}, true, true, 0},
		{"fee", []byte{0x3a, 2, 8, 42}, true, true, 42},
		{"merged", []byte{0x3a, 2, 8, 42, 0x3a, 0}, true, true, 42},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := hyperpb.NewMessage(typeOf)
			if err := msg.Unmarshal(tc.wire); err != nil {
				t.Fatal(err)
			}
			ref := dynamicpb.NewMessage(desc)
			if err := proto.Unmarshal(tc.wire, ref); err != nil {
				t.Fatal(err)
			}
			for _, m := range []protoreflect.Message{msg, ref} {
				if got := hyperpb.HasByIndex(m, 0); got != tc.config {
					t.Fatalf("config present=%v", got)
				}
				if hyperpb.HasByIndex(m, -1) || hyperpb.HasByIndex(m, 100) {
					t.Fatal("invalid index present")
				}
				if tc.config {
					c := hyperpb.MessageByIndex(m, 0)
					if hyperpb.HasByIndex(c, 0) != tc.fee || hyperpb.Uint64ByIndex(c, 0) != tc.value {
						t.Fatalf("fee presence/value: %v / %d", hyperpb.HasByIndex(c, 0), hyperpb.Uint64ByIndex(c, 0))
					}
				}
			}
		})
	}
	if hyperpb.HasByIndex(nil, 0) {
		t.Fatal("nil present")
	}
}
