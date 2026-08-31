# Counterexample memory protocol v1

The memory boundary is a read-only projection over two inputs: an append-only
NDJSON evidence corpus and a candidate's replay claims. The evaluator never
repairs either input. A repair is a new record in a later corpus version.

## Record identity

Every evidence record has a stable `record_id`, an ordinal, a
`previous_record_digest`, and a `record_digest`. The digest is SHA-256 over the
canonical JSON record with only `record_digest` blanked. The previous digest
must equal the preceding line's digest. Line order, identity, and digest form
the append-only boundary.

Each record also preserves `corpus_lineage`, a digest-bound `replay` comparison,
a minimal `causal_frontier`, and `inherits_from` links. A later candidate must
refer to the same record ID and digest before it can claim retention or
resolution.

## Candidate classification

The evaluator emits one status per controlled case:

| Status | Meaning |
| --- | --- |
| `DETECTED` | The candidate replays the recorded counterexample and keeps the exact memory link. |
| `RESOLVED_WITH_EVIDENCE` | The candidate no longer reproduces the failure and exact resolution evidence, replay, frontier, and inheritance all bind to the record. |
| `UNKNOWN` | The candidate cannot be decided; all six unknown coordinates are mandatory. |
| `REFUTED` | A contradiction, digest mismatch, missing memory link, conflict, or denominator drift is observed. |

State aggregation uses `REFUTED > UNKNOWN > CLOSED`. In particular, an
evidence conflict or a known missing memory link is `REFUTED` even if another
field is missing; it is never laundered into `UNKNOWN`.

`DIRECT_MISSING` has an empty `blocked_by` list. `DEPENDENCY_BLOCKED` has at
least one predecessor or case ID in `blocked_by`. Both still carry
`stage`, `step`, `reason`, `unknown_class`, and `next_operation`.

## Denominator and external input

The 12-cell denominator is fixed for v1. A denominator digest change is itself
a controlled counterexample and is fail-closed. External release locks may be
passed as optional case input, but no external repository or other repository's
CI is a required gate.

