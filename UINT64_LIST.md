# Allocation-free borrowed uint64 lists

Included in fork release `v0.1.6`.

`GetUint64ListByIndexUnchecked` now returns the arena-owned `*Scalars` descriptor
through the existing `Uint64List` interface. Previously returning `*list` copied
the 16-byte descriptor into an escaping interface box. Yellowstone cached two
such boxes per parsed transaction: pre-balances and post-balances.

Public signatures, element values, field decoding, and the generic protobuf
fallback are unchanged. Hyperpb lists borrow the message arena and any aliased
input bytes. Consume them before `Shared.Free` / Yellowstone view `Release`;
never retain them in an asynchronous callback. This is not an owned snapshot.
Yellowstone's existing view reset clears both cached balance lists before reuse.

## Validation

`go test ./...` passes. `TestUint64ListBorrowedZeroAllocAndReuse` covers uint64
and fixed64, packed/unpacked encoding, aliasing enabled/disabled, zero-copy byte
values and larger arena-decoded values, empty/growing/shrinking lists, `Get`,
scratch-backed `Copy`, invalid indexes, repeated arena Free/reuse, and an
interface-escaping allocation counter.

The libshreder consumer also has:

- `TestBalanceViewsAcrossReleaseAndReuse`, including double Release;
- `TestWarmedDecodeAndParseZeroAlloc`, enabled with `-tags hyperpb_zeroalloc`.

The latter is an opt-in release gate because the consumer still pins published
hyperpb v0.1.5, which allocates. Run it with a development `-modfile` mapping
`buf.build/go/hyperpb` to this working tree, or upgrade that consumer's fork
replacement to v0.1.6. Do not deploy a local filesystem replacement.

## Local benchmark, 2026-09-08

Go 1.26.0, darwin/arm64, Apple M5 Pro. Three alternating before/after executable
runs, 300ms per case. Medians of benchmark means, not production p50 latency:

| Decode + parse | Before ns/op | After ns/op |
|---|---:|---:|
| Pump Legacy | 958.9 | 938.0 |
| Pump V0 | 968.4 | 945.5 |
| Pump V1 | 1105 | 1071 |
| PumpAMM Legacy | 1243 | 1233 |
| PumpAMM V0 | 1254 | 1235 |
| PumpAMM V1 | 1337 | 1321 |

All six cases change from 32 B/op, 2 allocs/op to 0 B/op, 0 allocs/op.
The small timing change is secondary to the allocation reduction. This result
does not promise zero allocation for transport setup, larger scratch growth,
owned protobuf conversion, every protocol handler, or generic fallback lists.
