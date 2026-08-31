# RFC: Gooo counterexample memory v1

## Problem

A self-improving language can appear to improve while silently dropping a
failure discovered by an earlier version. A numeric score cannot prove that a
specific counterexample survived, was replayed, or was resolved with evidence.

## Proposal

The language's meta-program declares twelve activities that create and connect
counterexample declarations, immutable records, digests, corpus links, replay
comparisons, causal frontiers, resolution evidence, conflicts, denominator
changes, and inheritance. The Go projection is only an evaluator of those
bindings; it cannot invent an activity that is absent from the `.gooo` source.

The controlled corpus contains twelve records and twelve candidate cases. The
CI artifact reports exact integer counts for retained, forgotten, resolved, and
unknown cases. It does not compute an aggregate score or percentage.

## Compatibility rule

Existing records are never rewritten. To change the corpus, a new corpus
version must point to the prior corpus digest and append new evidence. A
candidate may claim `RESOLVED_WITH_EVIDENCE` only with an exact predecessor
record digest, a matching replay comparison, the recorded causal frontier, the
record's inheritance links, and the expected resolution evidence.

## Release rule

v0.1.0 is a tagged immutable source release. The release workflow creates a
source archive, SHA-256 checksum file, and manifest binding the asset to the
tag commit. It does not overwrite an existing release and does not depend on
any other repository's CI.

