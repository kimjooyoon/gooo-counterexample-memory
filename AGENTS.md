# Contribution contract

owner_thread_id: 01a0558a-d0a2-7b30-9f50-dc21dbb6f047
scope: gooo-counterexample-memory
shared_workspace_policy: single-writer

This repository is an evidence evaluator, not a text-only report generator.

- `.gooo` declarations are the authority for the 12 meta activities.
- `contracts/counterexample-memory-denominator-v1.json` is an immutable fixed
  denominator. It may only be superseded by a separately named version with a
  migration record; a candidate cannot change it during evaluation.
- Evidence records are append-only. Their sequence, `previous_digest`, and
  `record_digest` form a hash chain. Existing records are never rewritten.
- A candidate may resolve a counterexample only with exact evidence bound to
  the predecessor record, candidate digest, replay comparison, and causal
  frontier.
- `REFUTED` takes precedence over `UNKNOWN`, and `UNKNOWN` takes precedence over
  `CLOSED`. An observed contradiction is never hidden by missing evidence.
- `UNKNOWN` requires all six fields: `stage`, `step`, `reason`,
  `unknown_class`, `next_operation`, and `blocked_by`. `DIRECT_MISSING` and
  `DEPENDENCY_BLOCKED` are distinct classes.
- CI is the only Go verification boundary. Do not run local Go build, test,
  vet, or formatting validation as part of the conformance procedure.
- The evaluator is read-only with respect to the repository. Reports belong in
  a caller-owned temporary directory.

The current release workflow is intentionally independent of other repositories
and consumes only optional immutable external release locks.
