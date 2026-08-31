# v0.1.0 release checklist

The release is created only after the `Counterexample memory conformance`
workflow succeeds for the merged `main` commit.

The immutable release contains:

- a source archive named `gooo-counterexample-memory-v0.1.0.tar.gz`;
- `SHA256SUMS` with the archive digest; and
- `release-manifest.json` binding the archive digest to the tag commit.

The digest in the release manifest is the release asset identity to report to
consumers. External release locks are optional evaluator inputs and are not
required to pass this repository's CI.

