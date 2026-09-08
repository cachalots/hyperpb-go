# Indexed optional-field presence

`HasByIndex(msg, index)` provides protobuf presence checks for a known descriptor
index, with a fast path on hyperpb messages and a generated/dynamic protobuf
fallback. Nil messages and out-of-range indexes return false.

`Message.HasByIndexUnchecked(index)` is for callers caching a descriptor index
from the same message type; do not use untrusted field numbers as indexes.
Explicit optional scalar zero and an empty nested message both remain present.

This additive API does not change wire decoding. It supports Yellowstone V1
config consumers without forcing an owned protobuf conversion. The regression
test covers absence, empty config, explicit zero, nonzero, repeated-message
merge, and generic protobuf fallback.

```
go test . -run 'TestHasByIndex|TestProtoReset|TestProtoUnmarshal' -count=1
```

Included in fork release `v0.1.6`, together with allocation-free borrowed uint64
list descriptors. See `UINT64_LIST.md` for measurements and lifetime rules.
