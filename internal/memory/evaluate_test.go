package memory

import "testing"

func TestRecordDigestChangesWhenEvidenceChanges(t *testing.T) {
	record := EvidenceRecord{
		Schema:                       CorpusSchema,
		RecordID:                     "ce-test",
		Ordinal:                     1,
		Kind:                         "CONTRADICTION",
		Subject:                      "test://subject",
		SourceVersion:                "v0.1.0",
		ObservedCounterexampleDigest: "sha256:observed",
		CorpusLineage: CorpusLineage{
			CorpusID:           "test-corpus",
			ParentCorpusDigest: "sha256:parent",
			SourceVersion:      "v0.1.0",
			LineageDigest:      "sha256:lineage",
		},
		Replay: ReplayComparison{
			PreviousVersion:        "v0.0.1",
			CandidateVersion:       "v0.1.0",
			InputDigest:            "sha256:input",
			PreviousOutcomeDigest:  "sha256:bad",
			CandidateOutcomeDigest: "sha256:bad",
			Difference:             "reproduced",
			ComparisonDigest:       "sha256:replay",
		},
		CausalFrontier:       []string{"decl:test"},
		InheritsFrom:         []string{},
		EvidenceConflict:     nil,
		DenominatorDigest:    "sha256:denominator",
		PreviousRecordDigest: "",
	}
	left := DigestRecord(record)
	record.ObservedCounterexampleDigest = "sha256:changed"
	if left == DigestRecord(record) {
		t.Fatal("record digest did not change with immutable evidence")
	}
}

func TestUnknownClassesKeepDirectAndDependencySeparate(t *testing.T) {
	if !validUnknown(&UnknownTuple{Stage: "S", Step: "T", Reason: "R", UnknownClass: UnknownDirectMissing, NextOperation: "N", BlockedBy: []string{}}) {
		t.Fatal("direct missing tuple should be valid with no dependency")
	}
	if validUnknown(&UnknownTuple{Stage: "S", Step: "T", Reason: "R", UnknownClass: UnknownDirectMissing, NextOperation: "N", BlockedBy: []string{"dependency"}}) {
		t.Fatal("direct missing tuple must not carry dependency blockers")
	}
	if !validUnknown(&UnknownTuple{Stage: "S", Step: "T", Reason: "R", UnknownClass: UnknownDependency, NextOperation: "N", BlockedBy: []string{"case:previous"}}) {
		t.Fatal("dependency blocked tuple should name its blocker")
	}
}

func TestKnownConflictRefutesBeforeUnknown(t *testing.T) {
	record := EvidenceRecord{
		Schema:                       CorpusSchema,
		RecordID:                     "ce-conflict",
		Ordinal:                     1,
		Kind:                         "EVIDENCE_CONFLICT",
		Subject:                      "test://conflict",
		SourceVersion:                "v0.1.0",
		ObservedCounterexampleDigest: "sha256:observed",
		CorpusLineage:               CorpusLineage{CorpusID: "test", ParentCorpusDigest: "sha256:p", SourceVersion: "v0.1.0", LineageDigest: "sha256:l"},
		Replay:                      ReplayComparison{PreviousVersion: "v0.0.1", CandidateVersion: "v0.1.0", InputDigest: "sha256:i", PreviousOutcomeDigest: "sha256:b", CandidateOutcomeDigest: "sha256:b", Difference: "conflict", ComparisonDigest: "sha256:r"},
		CausalFrontier:              []string{"decl:test"},
		InheritsFrom:                []string{},
		EvidenceConflict:            &EvidenceConflict{LeftDigest: "sha256:left", RightDigest: "sha256:right", ConflictReason: "conflict"},
		DenominatorDigest:           "sha256:d",
	}
	input := CaseInput{Schema: InputSchema, CaseID: "ce-conflict", Kind: "EVIDENCE_CONFLICT", EvidenceRecordID: "ce-conflict", EvidenceRecordDigest: "sha256:record", CandidateMemoryRecordID: "ce-conflict", CandidateClaim: StatusUnknown}
	result := evaluateCase(testMeta(), map[string]EvidenceRecord{"ce-conflict": record}, input)
	if result.Status != StatusRefuted {
		t.Fatalf("known conflict was lowered to %s", result.Status)
	}
}

func testMeta() Meta {
	bindings := make(map[string]ActivityBinding, 12)
	activities := []string{
		"DeclareCounterexample", "AppendImmutableEvidenceRecord", "DigestEvidenceRecord", "LinkCounterexampleToCorpus",
		"ReplayPriorVersion", "CompareReplayOutcome", "EmitCausalFrontier", "AttachResolutionEvidence",
		"ResolveCounterexampleWithEvidence", "PreserveEvidenceConflict", "ValidateDenominatorChange", "InheritCounterexampleAcrossVersion",
	}
	for _, activity := range activities {
		bindings[activity] = ActivityBinding{Activity: activity, SourcePath: "main.gooo", SourceDigest: "sha256:source", IRNode: "ir:" + activity, GeneratedArtifact: "evaluation.json", Evaluator: "scripts/evaluate.sh", MetricID: "metric:" + activity}
	}
	return Meta{SourcePath: "main.gooo", SourceDigest: "sha256:source", ContractDigest: "sha256:contract", IRDigest: "sha256:ir", Bindings: bindings}
}

