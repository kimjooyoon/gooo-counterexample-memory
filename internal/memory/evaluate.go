package memory

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func LoadCorpus(path string) ([]EvidenceRecord, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var records []EvidenceRecord
	var previous string
	for scanner.Scan() {
		var record EvidenceRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, "", fmt.Errorf("decode corpus record: %w", err)
		}
		if record.Schema != CorpusSchema || record.Ordinal != len(records)+1 || record.RecordID == "" || record.RecordDigest == "" {
			return nil, "", errors.New("corpus record is not a valid append-only entry")
		}
		if record.PreviousRecordDigest != previous || DigestRecord(record) != record.RecordDigest {
			return nil, "", fmt.Errorf("corpus digest chain failed at %s", record.RecordID)
		}
		if record.CorpusLineage.LineageDigest != DigestLineage(record.CorpusLineage) {
			return nil, "", fmt.Errorf("corpus lineage digest failed at %s", record.RecordID)
		}
		if record.Replay.ComparisonDigest != DigestReplay(record.Replay) {
			return nil, "", fmt.Errorf("replay comparison digest failed at %s", record.RecordID)
		}
		if record.ExpectedResolutionEvidence != nil && record.ExpectedResolutionEvidence.EvidenceID == "" {
			return nil, "", fmt.Errorf("resolution evidence is incomplete at %s", record.RecordID)
		}
		records = append(records, record)
		previous = record.RecordDigest
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	if len(records) < 12 {
		return nil, "", fmt.Errorf("corpus has %d records; at least 12 are required", len(records))
	}
	return records, DigestCorpus(records), nil
}

func DigestCorpus(records []EvidenceRecord) string {
	data, err := json.Marshal(records)
	if err != nil {
		return Digest([]byte(fmt.Sprintf("%v", records)))
	}
	return Digest(data)
}

func LoadCases(path string) ([]CaseInput, error) {
	paths, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	if len(paths) != 12 {
		return nil, fmt.Errorf("controlled corpus requires exactly 12 case files, found %d", len(paths))
	}
	cases := make([]CaseInput, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var input CaseInput
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("decode case %s: %w", path, err)
		}
		if input.Schema != InputSchema || input.CaseID == "" || seen[input.CaseID] {
			return nil, fmt.Errorf("invalid or duplicate case %s", path)
		}
		seen[input.CaseID] = true
		cases = append(cases, input)
	}
	return cases, nil
}

func Evaluate(meta Meta, corpusPath string, records []EvidenceRecord, corpusDigest string, cases []CaseInput) EvaluationReport {
	byID := make(map[string]EvidenceRecord, len(records))
	for _, record := range records {
		byID[record.RecordID] = record
	}
	report := EvaluationReport{
		Schema:                        Schema,
		Decision:                      "FAIL_CLOSED",
		State:                         StateRefuted,
		Precedence:                   append([]string(nil), Precedence...),
		SourcePath:                    meta.SourcePath,
		SourceDigest:                  meta.SourceDigest,
		ContractDigest:                meta.ContractDigest,
		SemanticIRDigest:              meta.IRDigest,
		CorpusPath:                    corpusPath,
		CorpusDigest:                  corpusDigest,
		CorpusRecordCount:             len(records),
		AppendOnly:                    true,
		ExternalReleaseInputsOptional: true,
		CrossProjectRequiredGates:     0,
		Cases:                         make([]CaseResult, 0, len(cases)),
	}
	usedActivities := map[string]bool{}
	anyUnknown := false
	anyRefuted := false
	for _, input := range cases {
		result := evaluateCase(meta, byID, input)
		report.Cases = append(report.Cases, result)
		switch result.Status {
		case StatusDetected:
			report.Summary.Retained++
		case StatusResolved:
			report.Summary.Resolved++
		case StatusUnknown:
			report.Summary.Unknown++
			anyUnknown = true
		case StatusRefuted:
			report.Summary.Forgotten++
			anyRefuted = true
		}
		for _, binding := range result.Activities {
			usedActivities[binding.Activity] = true
		}
	}
	report.Summary.Total = len(report.Cases)
	for activity, binding := range meta.Bindings {
		if !usedActivities[activity] {
			report.AppendOnly = false
		}
		if binding.Activity != "" {
			report.ActivityBindings = append(report.ActivityBindings, binding)
		}
	}
	sort.Slice(report.ActivityBindings, func(i, j int) bool { return report.ActivityBindings[i].Activity < report.ActivityBindings[j].Activity })
	if anyRefuted {
		report.State = StateRefuted
		report.Decision = "FAIL_CLOSED"
	} else if anyUnknown {
		report.State = StateUnknown
		report.Decision = "COUNTEREXAMPLE_MEMORY_UNKNOWN"
	} else if len(report.ActivityBindings) == 12 && report.Summary.Total == 12 {
		report.State = StateClosed
		report.Decision = "COUNTEREXAMPLE_MEMORY_CLOSED"
	}
	return report
}

func evaluateCase(meta Meta, byID map[string]EvidenceRecord, input CaseInput) CaseResult {
	result := CaseResult{
		Schema:               Schema,
		CaseID:               input.CaseID,
		Kind:                 input.Kind,
		Subject:              input.Subject,
		Status:               StatusRefuted,
		State:                StateRefuted,
		Retention:            "FORGOTTEN",
		EvidenceRecordID:     input.EvidenceRecordID,
		EvidenceRecordDigest: input.EvidenceRecordDigest,
		ExternalRelease:      input.ExternalRelease,
	}
	record, ok := byID[input.EvidenceRecordID]
	if !ok || input.EvidenceRecordDigest != record.RecordDigest {
		return withActivities(meta, result, input, "IMMUTABLE_EVIDENCE_RECORD_MISSING_OR_DIGEST_MISMATCH", false)
	}
	if input.Kind != record.Kind || input.CandidateVersion == "" || input.CandidateVersion != record.SourceVersion {
		return withActivities(meta, result, input, "COUNTEREXAMPLE_DECLARATION_MISMATCH", false)
	}
	if input.CandidateMemoryRecordID != record.RecordID {
		return withActivities(meta, result, input, "COUNTEREXAMPLE_FORGOTTEN_BY_CANDIDATE", false)
	}
	result.CausalFrontier = append([]string(nil), input.CausalFrontier...)
	result.Replay = record.Replay
	result.InheritsFrom = append([]string(nil), input.InheritsFrom...)
	if input.CandidateClaim == StatusUnknown && record.EvidenceConflict != nil {
		return withActivities(meta, result, input, "EVIDENCE_CONFLICT_REFUTED_BEFORE_UNKNOWN", false)
	}
	if input.CandidateClaim == StatusUnknown && record.Kind == "DENOMINATOR_CHANGE" && input.CandidateDenominatorDigest != record.DenominatorDigest {
		return withActivities(meta, result, input, "DENOMINATOR_CHANGE_REFUTED_BEFORE_UNKNOWN", false)
	}
	if record.Kind == "CORPUS_LINEAGE" && input.EvidenceRecordDigest == record.RecordDigest && input.Kind == record.Kind && input.CandidateObservationDigest != record.ObservedCounterexampleDigest {
		return withActivities(meta, result, input, "CORPUS_LINEAGE_DIGEST_DRIFT_REFUTED", false)
	}
	if record.Kind == "REPLAY_COMPARISON" {
		if input.Replay == nil {
			return unknownResult(meta, result, input, UnknownTuple{Stage: "REPLAY", Step: "COMPARE_REPLAY_OUTCOME", Reason: "REPLAY_COMPARISON_MISSING", UnknownClass: UnknownDirectMissing, NextOperation: "PROVIDE_REPLAY_COMPARISON", BlockedBy: []string{}})
		}
	}
	if input.Replay != nil && (input.Replay.ComparisonDigest != DigestReplay(*input.Replay) || CanonicalDigest(input.Replay) != CanonicalDigest(record.Replay)) {
		if input.CandidateClaim == StatusUnknown && record.EvidenceConflict != nil {
			return withActivities(meta, result, input, "EVIDENCE_CONFLICT_REFUTED_BEFORE_UNKNOWN", false)
		}
		if record.Kind == "REPLAY_COMPARISON" || input.CandidateClaim != StatusUnknown {
			return withActivities(meta, result, input, "REPLAY_COMPARISON_CONTRADICTED", false)
		}
	}
	if record.EvidenceConflict != nil {
		return withActivities(meta, result, input, "EVIDENCE_CONFLICT_NOT_RESOLVED", false)
	}
	if record.Kind == "DENOMINATOR_CHANGE" && input.CandidateDenominatorDigest != record.DenominatorDigest {
		return withActivities(meta, result, input, "DENOMINATOR_CHANGE_NOT_BOUND", false)
	}
	if input.CandidateClaim == StatusUnknown {
		if !validUnknown(input.Unknown) {
			return withActivities(meta, result, input, "UNKNOWN_TUPLE_INCOMPLETE", false)
		}
		return unknownResult(meta, result, input, *input.Unknown)
	}
	if input.CandidateClaim == StatusResolved {
		if input.ResolutionEvidence == nil || record.ExpectedResolutionEvidence == nil || CanonicalDigest(input.ResolutionEvidence) != CanonicalDigest(record.ExpectedResolutionEvidence) {
			return unknownResult(meta, result, input, UnknownTuple{Stage: "RESOLUTION", Step: "ATTACH_RESOLUTION_EVIDENCE", Reason: "RESOLUTION_EVIDENCE_MISSING", UnknownClass: UnknownDirectMissing, NextOperation: "PROVIDE_RESOLUTION_EVIDENCE", BlockedBy: []string{}})
		}
		if CanonicalDigest(input.CausalFrontier) != CanonicalDigest(record.CausalFrontier) || CanonicalDigest(input.InheritsFrom) != CanonicalDigest(record.InheritsFrom) {
			return withActivities(meta, result, input, "RESOLUTION_FRONTIER_OR_INHERITANCE_MISMATCH", false)
		}
		result.Status = StatusResolved
		result.State = StateClosed
		result.Retention = "RESOLVED"
		result.Reason = "EXACT_RESOLUTION_EVIDENCE_BOUND"
		result.ResolutionEvidence = input.ResolutionEvidence
		return withActivities(meta, result, input, result.Reason, true)
	}
	if input.CandidateClaim == StatusDetected && input.CandidateObservationDigest == record.ObservedCounterexampleDigest && CanonicalDigest(input.CausalFrontier) == CanonicalDigest(record.CausalFrontier) && CanonicalDigest(input.InheritsFrom) == CanonicalDigest(record.InheritsFrom) {
		result.Status = StatusDetected
		result.State = StateClosed
		result.Retention = "RETAINED"
		result.Reason = "COUNTEREXAMPLE_REPLAY_DETECTED_AND_RETAINED"
		return withActivities(meta, result, input, result.Reason, false)
	}
	return withActivities(meta, result, input, "CANDIDATE_OUTCOME_REFUTED", false)
}

func validUnknown(unknown *UnknownTuple) bool {
	if unknown == nil || unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" {
		return false
	}
	if unknown.UnknownClass == UnknownDirectMissing {
		return len(unknown.BlockedBy) == 0
	}
	if unknown.UnknownClass == UnknownDependency {
		return len(unknown.BlockedBy) > 0
	}
	return false
}

func unknownResult(meta Meta, result CaseResult, input CaseInput, unknown UnknownTuple) CaseResult {
	result.Status = StatusUnknown
	result.State = StateUnknown
	result.Retention = "UNKNOWN"
	result.Reason = unknown.Reason
	result.Unknown = &unknown
	return withActivities(meta, result, input, unknown.Reason, true)
}

func withActivities(meta Meta, result CaseResult, input CaseInput, reason string, evidenceBound bool) CaseResult {
	result.Reason = reason
	activities := []string{"DeclareCounterexample", "AppendImmutableEvidenceRecord", "DigestEvidenceRecord", "LinkCounterexampleToCorpus", "ReplayPriorVersion", "CompareReplayOutcome", "EmitCausalFrontier", "InheritCounterexampleAcrossVersion"}
	if evidenceBound {
		activities = append(activities, "AttachResolutionEvidence", "ResolveCounterexampleWithEvidence")
	}
	if result.Kind == "EVIDENCE_CONFLICT" {
		activities = append(activities, "PreserveEvidenceConflict")
	}
	if result.Kind == "DENOMINATOR_CHANGE" {
		activities = append(activities, "ValidateDenominatorChange")
	}
	for _, activity := range activities {
		if binding, ok := meta.Bindings[activity]; ok {
			result.Activities = append(result.Activities, binding)
		}
	}
	return result
}
