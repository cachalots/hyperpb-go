package hyperpb_test

import (
	"testing"

	"buf.build/go/hyperpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProtoResetZeroMessageDoesNotPanic(t *testing.T) {
	var msg hyperpb.Message

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("proto.Reset panicked: %v", r)
		}
	}()

	proto.Reset(&msg)
}

func TestProtoUnmarshalIntoDynamicMessageDoesNotPanic(t *testing.T) {
	msgType := hyperpb.CompileMessageDescriptor((&emptypb.Empty{}).ProtoReflect().Descriptor())
	msg := hyperpb.NewMessage(msgType)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("proto.Unmarshal panicked: %v", r)
		}
	}()

	if err := proto.Unmarshal(nil, msg); err != nil {
		t.Fatalf("proto.Unmarshal returned error: %v", err)
	}
}
