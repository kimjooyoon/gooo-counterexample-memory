# gooo-counterexample-memory

`gooo-counterexample-memory` is a small, read-only evaluator for the part of a
self-improving language that must not be forgotten: a counterexample discovered
by one version is carried into the next version as an immutable evidence record
and digest.

The evaluator reads a fixed 12-cell contract, an append-only counterexample
ledger, and a candidate input. It replays every ledger record against the
candidate and assigns exactly one of these states to every controlled case:

- `DETECTED`: the candidate still reproduces the counterexample; the evidence
  is retained.
- `RESOLVED_WITH_EVIDENCE`: the candidate no longer reproduces it and an exact,
  digest-bound resolution evidence record proves the change.
- `UNKNOWN`: the candidate cannot be decided. The record includes all six
  coordinates: `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
  `blocked_by`.
- `REFUTED`: a contradiction, tampered digest, denominator change, malformed
  evidence, or other fail-closed condition invalidates the candidate result.

At the report level the precedence is `REFUTED > UNKNOWN > CLOSED`; a report is
`CLOSED` only when all cases are either `DETECTED` or
`RESOLVED_WITH_EVIDENCE`. A known contradiction is never converted to
`UNKNOWN`.

## Fixed denominator

The denominator is exactly 12 cells, each with one released `.gooo` activity,
source binding, semantic IR binding, generated artifact, and evaluator. The
activities directly create or connect declaration, replay, causal frontier,
resolution evidence, and inheritance records. The denominator is not inferred
from file repetition.

The controlled fixture has 12 cases and deliberately includes corpus
lineage/digest drift, replay comparison drift, evidence conflict, denominator
change, direct missing evidence, dependency blocking, inheritance, and digest
tampering. The CI summary and `summary.json` report integer exact counts only:
`retained`, `forgotten`, `resolved`, and `unknown`. No score or percentage is
computed. `forgotten` is the fail-closed `REFUTED` bucket: it means the
candidate could not prove that the predecessor counterexample was retained or
resolved, never that the immutable ledger entry was deleted.

## Authority boundary

The evaluator never edits its source repository. It writes reports only under a
caller-owned output directory. `repository_writes`, `local_test_executions`,
and `cross_project_required_gates` are explicit integer fields in the CI
artifact and are all `0`. The root `README.md` is a documentation inventory
exception and is excluded from the physical inventory counts.

External release locks are optional inputs. When supplied, they must identify
an immutable tag, commit, and SHA-256 asset digest; no live external repository
or another repository's CI is a required gate.

## Release provenance

The GitHub API audit found that the existing `v0.1.0` release has
`immutable=false`. It is preserved as `REFUTED_RELEASE_IMMUTABILITY` and is not
counted as successful release evidence. The repository immutable-releases
setting is now enabled; the next release must pass both REST and GraphQL
`immutable=true` checks. See
[provenance/release-history-v1.json](provenance/release-history-v1.json).

## CI-only verification

GitHub Actions is the verification authority and uses Go 1.27. The repository
intentionally does not ask contributors to run local Go build, test, format, or
vet commands. The workflow runs conformance, formatting, tests, race tests,
and vet in CI, and uploads a machine-readable artifact containing the exact
case states and count fields.

See [docs/protocol-v1.md](docs/protocol-v1.md) for the state machine and
[docs/rfc-v1.md](docs/rfc-v1.md) for the evidence and inheritance contract.
